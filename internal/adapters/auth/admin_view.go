package authadapter

import (
	auth "github.com/go-sum/auth"
	authmodel "github.com/go-sum/auth/model"
	"github.com/go-sum/componentry/patterns/pager"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/forge/internal/view/partial/userpartial"

	g "maragu.dev/gomponents"
)

var _ auth.AdminPageRenderer = (*Renderer)(nil)

func (r *Renderer) AdminElevatePage(req auth.Request) g.Node {
	return page.AdminElevatePage(hostRequest(req))
}

func (r *Renderer) UserListPage(req auth.Request, data auth.AdminUsersPageData) g.Node {
	return page.UserListPage(hostRequest(req), page.UserListData{
		Users: data.Users,
		Pager: pagerFromAuth(data),
	})
}

func (r *Renderer) UserListRegion(req auth.Request, data auth.AdminUsersPageData) g.Node {
	return page.UserListRegion(hostRequest(req), page.UserListData{
		Users: data.Users,
		Pager: pagerFromAuth(data),
	})
}

func (r *Renderer) UserEditForm(req auth.Request, data auth.AdminUserFormData) g.Node {
	return userpartial.UserEditForm(hostRequest(req), userpartial.UserFormData{
		User:   data.User,
		Values: data.Values,
		Errors: data.Errors,
	})
}

func (r *Renderer) UserRow(req auth.Request, user authmodel.User) g.Node {
	return userpartial.UserRow(hostRequest(req), userpartial.UserRowProps{
		User: user,
	})
}

func hostRequest(req auth.Request) view.Request {
	if host, ok := req.State.(view.Request); ok {
		return host
	}
	return view.Request{
		CSRFToken:     req.CSRFToken,
		CSRFFieldName: req.CSRFFieldName,
	}
}

func pagerFromAuth(data auth.AdminUsersPageData) pager.Pager {
	return pager.Pager{
		Page:       data.Page,
		PerPage:    data.PerPage,
		TotalItems: data.TotalItems,
		TotalPages: data.TotalPages,
	}
}
