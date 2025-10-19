package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/Rahul4469/lenslocked/context"
	"github.com/Rahul4469/lenslocked/errors"
	"github.com/Rahul4469/lenslocked/models"
	"github.com/go-chi/chi/v5"
)

type Galleries struct {
	Template struct {
		Show  Template
		New   Template
		Edit  Template
		Index Template
	}
	GalleryService *models.GalleryService
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")
	g.Template.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID int
		Title  string
	}

	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Template.New.Execute(w, r, data, err)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	// id, err := strconv.Atoi(chi.URLParam(r, "id"))
	// if err != nil {
	// 	http.Error(w, "Invalid ID", http.StatusNotFound)
	// 	return
	// }
	// gallery, err := g.GalleryService.ByID(id)
	// if err != nil {
	// 	if errors.Is(err, models.ErrNotFound) {
	// 		http.Error(w, "Gallery not found", http.StatusNotFound)
	// 		return
	// 	}
	// 	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	// 	return
	// }

	//used helper func to retreive ID from URL & then
	//fetch gallery from DB using that ID
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string //Extra field added, more than that from model
	}
	var data struct {
		ID     int
		Title  string
		Images []Image
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}

	g.Template.Show.Execute(w, r, data)
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	//	used helper func to retreive ID from URL & then
	//	fetch gallery" from DB using that ID
	//	(then, also added a function to fetch user from context into the args of galleryByID)
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	// user := context.User(r.Context())
	// if gallery.UserID != user.ID {
	// 	http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
	// 	return
	// }

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string //Extra field added, more than that from model
	}
	var data struct {
		ID     int
		Title  string
		Images []Image
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	g.Template.Edit.Execute(w, r, data)

}
func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	// used helper func to retreive ID from URL & then
	// fetch gallery" from DB using that ID
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	err = userMustOwnGallery(w, r, gallery)
	if err != nil {
		return
	}
	// user := context.User(r.Context())
	// if gallery.UserID != user.ID {
	// 	http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
	// 	return
	// }

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)

}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
	}
	//data to render the galleries - slice of Gallery{}
	var data struct {
		Galleries []Gallery
	}

	user := context.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID:    gallery.ID,
			Title: gallery.Title,
		})
	}

	g.Template.Index.Execute(w, r, data)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	err = g.GalleryService.Delete(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid Gallery ID", http.StatusNotFound)
		return
	}
	image, err := g.GalleryService.Image(galleryID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	// var requestedImage models.Image
	// imageFound := false
	// for _, image := range images {
	// 	if image.Filename == filename {
	// 		requestedImage = image
	// 		imageFound = true
	// 		break
	// 	}
	// }
	// if !imageFound {
	// 	http.Error(w, "Image not Found", http.StatusNotFound)
	// 	return
	// }
	http.ServeFile(w, r, image.Path)
}

func (g Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = r.ParseMultipartForm(5 << 20) // 5mb: equivalent to (5 * 1024 * 1024)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			var fileError models.FileError
			if errors.As(err, &fileError) {
				msg := fmt.Sprintf(`%v has an invalid content type or extensions, 
							Only png, gif, anf jpg files can be uploaded.`, fileHeader.Filename)
				http.Error(w, msg, http.StatusBadRequest)
			}
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)

}

func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

//----------------------------------------------------
// Helper functions

// To make sure the code only uses the base name of the image from path
// so that longer path cant be injected into our code to access files
// we dont want to be accessed by unauthorized users
func (g Galleries) filename(w http.ResponseWriter, r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}

// This is a function type,just like http.HandlerFunc but with an additional parameter
// for the Gallery model and an error return value.
// Methods can be attached to this type which uses extra parameters.
type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (g Galleries) galleryByID(w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, err
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}
	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}

	return gallery, nil
}

func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}
