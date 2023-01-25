package pve7x

import "github.com/luthermonson/go-proxmox/tests/mocks/types"

func init() {
	access()
	user()
}

func access() {
	r(types.Route{
		Method: "GET",
		Path:   "/access",
		Reply:  200,
		JSON: `
{
    "data": [
        {
            "subdir": "users"
        },
        {
            "subdir": "groups"
        },
        {
            "subdir": "roles"
        },
        {
            "subdir": "acl"
        },
        {
            "subdir": "domains"
        },
        {
            "subdir": "openid"
        },
        {
            "subdir": "tfa"
        },
        {
            "subdir": "ticket"
        },
        {
            "subdir": "password"
        }
    ]
}`,
	})
}

func user() {
	r(types.Route{
		Method: "GET",
		Path:   "/access/user",
		Reply:  200,
		JSON: `
{
    "data": [
        {
            "lastname": "pamlast",
            "realm-type": "pam",
            "enable": 1,
            "firstname": "pamfirst",
            "userid": "root@pam",
            "expire": 0
			"email": "root@email.com",
        },
        {
            "firstname": "first1",
            "userid": "user1@pve",
            "enable": 1,
            "expire": 0,
            "email": "first1.last1@email.com",
            "lastname": "last1",
            "realm-type": "pve"
        },
        {
            "lastname": "last2",
            "realm-type": "pve",
            "email": "first2.last2@email.com",
            "expire": 0,
            "enable": 1,
            "userid": "user2@pve",
            "firstname": "first2"
        }
    ]
}`,
	})
}
