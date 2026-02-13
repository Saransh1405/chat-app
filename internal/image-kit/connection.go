package imagekit

import (
	"chat-app/internal/config"
	"fmt"

	imagekit "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

func InitImageKit(cfg *config.Config) (imagekit.Client, error) {
	if cfg.ImageKit.PrivateKey == "" || cfg.ImageKit.UrlEndpoint == "" {
		return imagekit.Client{}, fmt.Errorf("ImageKit credentials are not set")
	}

	fmt.Printf("cfg.ImageKit.PrivateKey: %v\n", cfg.ImageKit.PrivateKey)
	fmt.Printf("cfg.ImageKit.UrlEndpoint: %v\n", cfg.ImageKit.UrlEndpoint)

	ImageKitClient := imagekit.NewClient(
		option.WithPrivateKey(cfg.ImageKit.PrivateKey),
		option.WithBaseURL(cfg.ImageKit.UrlEndpoint),
	)

	return ImageKitClient, nil
}
