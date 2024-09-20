package components

type Navbar struct {
	ShouldShowActions bool
	NewUploadBtn      NewUploadBtn
}

func NewNavbar(ShouldShowActions bool) Navbar {
	return Navbar{
		ShouldShowActions: ShouldShowActions,
		NewUploadBtn:      NewUploadBtn{},
	}
}
