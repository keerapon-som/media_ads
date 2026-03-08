package domain

type MediaPublisher struct {
	objectLib ObjectLibraryInterface
}

func NewMediaPublisher(objectLibraryAPI ObjectLibraryInterface) *MediaPublisher {
	return &MediaPublisher{
		objectLib: objectLibraryAPI,
	}
}
