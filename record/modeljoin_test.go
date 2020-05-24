package record

import (
	"github.com/appist/appy/support"
)

type (
	Committee struct {
		Model       `masters:"primary" autoIncrement:"committee_id" primaryKeys:"committee_id"`
		CommitteeID int64  `db:"committee_id" faker:"-"`
		Name        string `db:"name"`
	}

	Member struct {
		Model    `masters:"primary" autoIncrement:"member_id" primaryKeys:"member_id"`
		MemberID int64  `db:"member_id"`
		Name     string `db:"name"`
	}
)

func (s *modelSuite) insertCommittees() {
	committees := []Committee{
		{Name: "John"},
		{Name: "Mary"},
		{Name: "Amelia"},
		{Name: "Joe"},
	}

	count, err := s.model(&committees).Create().Exec()
	s.Equal(4, len(committees))
	s.Equal(int64(4), count)
	s.Nil(err)
}

func (s *modelSuite) insertMembers() {
	members := []Member{
		{Name: "John"},
		{Name: "Jane"},
		{Name: "Mary"},
		{Name: "David"},
		{Name: "Amelia"},
	}

	count, err := s.model(&members).Create().Exec()
	s.Equal(5, len(members))
	s.Equal(int64(5), count)
	s.Nil(err)
}

func (s *modelSuite) TestJoin() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_join_with_"+adapter)
		s.insertCommittees()
		s.insertMembers()

		{
			member := Member{}
			count, err := s.model(&member).Join("INNER JOIN committees c ON c.name = members.name").Count().Exec()
			s.Equal(int64(3), count)
			s.Nil(err)

			members := []Member{}
			count, err = s.model(&members).Join("INNER JOIN committees c ON c.name = members.name").Count().Exec()
			s.Equal(int64(3), count)
			s.Nil(err)
		}

		{
			member := Member{}
			count, err := s.model(&member).Join("INNER JOIN committees c ON c.name = members.name").Where("member_id = ?", 1).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			members := []Member{}
			count, err = s.model(&members).Join("INNER JOIN committees c ON c.name = members.name").Where("member_id in (?)", []int64{1, 3}).Find().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)
		}

		{
			type innerJoin struct {
				CommitteeID int64  `db:"committee_id"`
				Committee   string `db:"committee"`
				MemberID    int64  `db:"member_id"`
				Member      string `db:"member"`
			}

			results := []innerJoin{}
			member := Member{}
			count, err := s.model(&member).Select("members.member_id, members.name AS member, c.committee_id, c.name AS committee").Join("INNER JOIN committees c ON c.name = members.name").Scan(&results).Exec()
			s.Equal(int(3), len(results))
			s.Equal(int64(3), count)
			s.Nil(err)

			results = []innerJoin{}
			members := []Member{}
			count, err = s.model(&members).Select("members.member_id, members.name AS member, c.committee_id, c.name AS committee").Join("INNER JOIN committees c ON c.name = members.name").Scan(&results).Exec()
			s.Equal(int(3), len(results))
			s.Equal(int64(3), count)
			s.Nil(err)
		}
	}
}
