package components

type Navbar struct {
	ShouldShowActions bool
	ShouldShowLogout  bool
	NewUploadBtn      LinkBtn
	LogoutBtn         LinkBtn
}

func NewNavbar(shouldShowActions bool, shouldShowLogout bool) Navbar {
	return Navbar{
		ShouldShowActions: shouldShowActions,
		ShouldShowLogout:  shouldShowLogout,
		NewUploadBtn: LinkBtn{
			Href: "/",
			Text: "Upload New File",
		},
		LogoutBtn: LinkBtn{
			Href: "/logout",
			Text: "Logout",
		},
	}
}
