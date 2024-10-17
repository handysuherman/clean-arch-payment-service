package unierror

import "errors"

var (
	ErrUnsupportedFileType                                = errors.New("error code: 47400")
	ErrMinioPutObject                                     = errors.New("error code: 47401")
	ErrJPGDecode                                          = errors.New("error code: 47402")
	ErrPNGDecode                                          = errors.New("error code: 47403")
	ErrUnableToCreateFile                                 = errors.New("error code: 47404")
	ErrUnableToPresetLossyEncoder                         = errors.New("error code: 47405")
	ErrWEBPEncode                                         = errors.New("error code: 47406")
	ErrFileStatRetrieval                                  = errors.New("error code: 47407")
	ErrCreateProductUploadFileMax3File                    = errors.New("error code: 47408")
	ErrUploadedFileExceedingLimitFileSize                 = errors.New("error code: 47409")
	ErrUploadedFileExceedingFileLimitPerUpload            = errors.New("error code: 47410")
	ErrUploadedFilePermittedType                          = errors.New("error code: 47411")
	ErrUploadedImageShouldNotBeEmpty                      = errors.New("error code: 47412")
	ErrUploadedImageShouldOnlyOneSelectedAsPromotionImage = errors.New("error code: 47412")
)
