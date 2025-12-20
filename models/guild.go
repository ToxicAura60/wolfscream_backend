package models

type Guild struct {
	Id 		   string	 `json:"id"`
	Name     string  `json:"name"`
	ImageUrl *string `json:"image_url"`  
}
