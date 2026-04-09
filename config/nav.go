package config

import uilayout "github.com/go-sum/componentry/ui/layout"

type NavConfig = uilayout.NavConfig
type NavbarBrand = uilayout.NavbarBrand
type NavSection = uilayout.NavSection
type NavItem = uilayout.NavItem

func defaultNav() NavConfig {
	return NavConfig{
		Brand: uilayout.NavbarBrand{
			Label: "Starter",
			Href:  "/",
		},
		Sections: []uilayout.NavSection{
			{
				Items: []uilayout.NavItem{
					{Label: "Home", Href: "/"},
					{
						Label:      "Packages",
						Visibility: "user",
						Items: []uilayout.NavItem{
							{
								Label: "Componentry",
								Items: []uilayout.NavItem{
									{Label: "Components", Href: "/_components"},
								},
							},
							{
								Label: "Auth",
								Items: []uilayout.NavItem{
									{Label: "Users", Href: "/admin/users"},
									{Label: "Sessions", Href: "/profile/sessions"},
								},
							},
							{
								Label: "Site",
								Items: []uilayout.NavItem{
									{Label: "Robots", Href: "/robots.txt"},
									{Label: "Sitemap", Href: "/sitemap.xml"},
								},
							},
						},
					},
				},
			},
			{
				Align: "end",
				Items: []uilayout.NavItem{
					{
						Label: "Features",
						Items: []uilayout.NavItem{
							{Label: "Documents", Href: "/docs"},
						},
					},
					{Label: "Contact us", Href: "/contact"},
					{
						Label: "Account",
						Items: []uilayout.NavItem{
							{Label: "Sign In", Href: "/signin", Visibility: "guest"},
							{Label: "Sign Up", Href: "/signup", Visibility: "guest"},
							{Slot: "user_name", Visibility: "user"},
							{Slot: "signout", Label: "Signout", Visibility: "user"},
						},
					},
					{Slot: "theme_toggle"},
				},
			},
		},
	}
}
