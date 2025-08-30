package main

import "github.com/cloudinary/cloudinary-go/v2"

type cloudinaryApi struct {
	cloudinaryUrl string
}

func newCloudinaryApi(cloudinaryUrl string) *cloudinaryApi {
	return &cloudinaryApi{
		cloudinaryUrl: cloudinaryUrl,
	}
}

func (c *cloudinaryApi) createInstance() (*cloudinary.Cloudinary, error) {

	cld, err := cloudinary.NewFromURL(c.cloudinaryUrl)
	if err != nil {
		return nil, err
	}

	return cld, nil
}
