package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// PhotoFactory creates test photo entities
type PhotoFactory struct{}

// NewPhotoFactory creates a new photo factory
func NewPhotoFactory() *PhotoFactory {
	return &PhotoFactory{}
}

// CreatePhoto creates a test photo with default values
func (f *PhotoFactory) CreatePhoto() *entities.Photo {
	now := time.Now()
	photoID := uuid.New()
	userID := uuid.New()
	
	return &entities.Photo{
		ID:          photoID,
		UserID:      userID,
		URL:         "https://example.com/photos/" + photoID.String() + ".jpg",
		ThumbnailURL: "https://example.com/photos/thumbnails/" + photoID.String() + ".jpg",
		IsPrimary:   false,
		IsApproved:  false,
		IsPrivate:   false,
		ViewCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreatePrimaryPhoto creates a primary test photo
func (f *PhotoFactory) CreatePrimaryPhoto() *entities.Photo {
	photo := f.CreatePhoto()
	photo.IsPrimary = true
	return photo
}

// CreateApprovedPhoto creates an approved test photo
func (f *PhotoFactory) CreateApprovedPhoto() *entities.Photo {
	photo := f.CreatePhoto()
	photo.IsApproved = true
	return photo
}

// CreatePrivatePhoto creates a private test photo
func (f *PhotoFactory) CreatePrivatePhoto() *entities.Photo {
	photo := f.CreatePhoto()
	photo.IsPrivate = true
	return photo
}

// CreateCustomPhoto creates a test photo with custom values
func (f *PhotoFactory) CreateCustomPhoto(opts ...PhotoOption) *entities.Photo {
	photo := f.CreatePhoto()
	
	for _, opt := range opts {
		opt(photo)
	}
	
	return photo
}

// PhotoOption defines a function type for customizing photo creation
type PhotoOption func(*entities.Photo)

// WithPhotoID sets the photo ID
func WithPhotoID(id uuid.UUID) PhotoOption {
	return func(p *entities.Photo) {
		p.ID = id
	}
}

// WithUserID sets the user ID
func WithUserID(userID uuid.UUID) PhotoOption {
	return func(p *entities.Photo) {
		p.UserID = userID
	}
}

// WithURL sets the photo URL
func WithURL(url string) PhotoOption {
	return func(p *entities.Photo) {
		p.URL = url
	}
}

// WithThumbnailURL sets the thumbnail URL
func WithThumbnailURL(thumbnailURL string) PhotoOption {
	return func(p *entities.Photo) {
		p.ThumbnailURL = thumbnailURL
	}
}

// WithPrimary sets the primary status
func WithPrimary(primary bool) PhotoOption {
	return func(p *entities.Photo) {
		p.IsPrimary = primary
	}
}

// WithApproved sets the approval status
func WithApproved(approved bool) PhotoOption {
	return func(p *entities.Photo) {
		p.IsApproved = approved
	}
}

// WithPrivate sets the private status
func WithPrivate(private bool) PhotoOption {
	return func(p *entities.Photo) {
		p.IsPrivate = private
	}
}

// WithViewCount sets the view count
func WithViewCount(viewCount int) PhotoOption {
	return func(p *entities.Photo) {
		p.ViewCount = viewCount
	}
}

// WithCreatedAt sets the creation time
func WithPhotoCreatedAt(createdAt time.Time) PhotoOption {
	return func(p *entities.Photo) {
		p.CreatedAt = createdAt
	}
}

// WithPhotoUpdatedAt sets the update time
func WithPhotoUpdatedAt(updatedAt time.Time) PhotoOption {
	return func(p *entities.Photo) {
		p.UpdatedAt = updatedAt
	}
}

// CreateMultiplePhotos creates multiple test photos
func (f *PhotoFactory) CreateMultiplePhotos(count int) []*entities.Photo {
	photos := make([]*entities.Photo, count)
	for i := 0; i < count; i++ {
		photos[i] = f.CreatePhoto()
	}
	return photos
}

// CreateMultipleCustomPhotos creates multiple test photos with custom options
func (f *PhotoFactory) CreateMultipleCustomPhotos(count int, opts ...PhotoOption) []*entities.Photo {
	photos := make([]*entities.Photo, count)
	for i := 0; i < count; i++ {
		photos[i] = f.CreateCustomPhoto(opts...)
	}
	return photos
}

// CreatePhotosForUser creates multiple photos for a specific user
func (f *PhotoFactory) CreatePhotosForUser(userID uuid.UUID, count int) []*entities.Photo {
	photos := make([]*entities.Photo, count)
	for i := 0; i < count; i++ {
		photo := f.CreatePhoto()
		photo.UserID = userID
		if i == 0 {
			photo.IsPrimary = true // First photo is primary
		}
		photos[i] = photo
	}
	return photos
}