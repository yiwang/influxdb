package cluster

import (
	"common"
	. "launchpad.net/gocheck"
)

type UserSuite struct{}

var _ = Suite(&UserSuite{})

var root common.User

func (self *UserSuite) SetUpSuite(c *C) {
	user := &clusterAdmin{CommonUser{"root", "", false, "root"}}
	c.Assert(user.ChangePassword("password"), IsNil)
	root = user
}

func (self *UserSuite) TestProperties(c *C) {
	u := clusterAdmin{CommonUser{Name: "root"}}
	c.Assert(u.IsClusterAdmin(), Equals, true)
	c.Assert(u.GetName(), Equals, "root")
	hash, err := HashPassword("foobar")
	c.Assert(err, IsNil)
	c.Assert(u.ChangePassword(string(hash)), IsNil)
	c.Assert(u.isValidPwd("foobar"), Equals, true)
	c.Assert(u.isValidPwd("password"), Equals, false)

	dbUser := dbUser{CommonUser{Name: "db_user"}, "db", nil, nil, true}
	c.Assert(dbUser.IsClusterAdmin(), Equals, false)
	c.Assert(dbUser.IsDbAdmin("db"), Equals, true)
	c.Assert(dbUser.GetName(), Equals, "db_user")
	hash, err = HashPassword("password")
	c.Assert(err, IsNil)
	c.Assert(dbUser.ChangePassword(string(hash)), IsNil)
	c.Assert(dbUser.isValidPwd("password"), Equals, true)
	c.Assert(dbUser.isValidPwd("password1"), Equals, false)
}
