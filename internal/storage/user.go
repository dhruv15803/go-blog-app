package storage

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	Id         int      `db:"id" json:"id"`
	Email      string   `db:"email" json:"email"`
	Username   *string  `db:"username" json:"username"`
	Password   string   `db:"password" json:"-"`
	Name       *string  `db:"name" json:"name"`
	ProfileImg *string  `db:"profile_img" json:"profile_img"`
	IsVerified bool     `db:"is_verified" json:"is_verified"`
	Role       UserRole `db:"role" json:"role"`
	CreatedAt  string   `db:"created_at" json:"created_at"`
	UpdatedAt  *string  `db:"updated_at" json:"updated_at"`
}

type UserInvitation struct {
	Token      string `db:"token" json:"token"`
	UserId     int    `db:"user_id" json:"user_id"`
	Expiration string `db:"expiration" json:"expiration"`
}

func (s *Storage) GetUserByEmail(email string) (*User, error) {

	var user User

	query := `SELECT id,email,username,password,
    name,profile_img,is_verified,role,created_at,updated_at 
	FROM users WHERE email=$1`

	row := s.db.QueryRowx(query, email)
	if err := row.StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetVerifiedUserByEmail(email string) (*User, error) {
	var user User

	query := `SELECT id,id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
FROM users WHERE email=$1 AND is_verified=true`

	row := s.db.QueryRowx(query, email)
	if err := row.StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) CreateUserAndInvite(email string, password string, token string, inviteExpiration time.Time) (*User, error) {

	var user User

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	var rollBackErr error

	defer func() {
		if rollBackErr != nil {
			tx.Rollback()
		}
	}()

	query := `INSERT INTO users(email,password) VALUES($1,$2) RETURNING 
id,email,username,password,name,profile_img,is_verified,role,created_at,updated_at`

	if rollBackErr = tx.QueryRowx(query, email, password).StructScan(&user); rollBackErr != nil {
		return nil, rollBackErr
	}

	invitationQuery := `INSERT INTO user_invitations(token,user_id,expiration) VALUES($1,$2,$3)`

	_, rollBackErr = tx.Exec(invitationQuery, token, user.Id, inviteExpiration)
	if rollBackErr != nil {
		return nil, rollBackErr
	}
	if rollBackErr = tx.Commit(); rollBackErr != nil {
		return nil, rollBackErr
	}

	return &user, nil
}

func (s *Storage) ActivateUser(token string) (*User, error) {

	var userInvite UserInvitation

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	var rollBackErr error

	defer func() {
		if rollBackErr != nil {
			tx.Rollback()
		}
	}()

	query := `SELECT token,user_id,expiration FROM user_invitations WHERE token=$1 AND expiration > $2`

	if rollBackErr = tx.QueryRowx(query, token, time.Now()).StructScan(&userInvite); rollBackErr != nil {
		return nil, rollBackErr
	}

	var activeUser User
	activeUserId := userInvite.UserId

	verifyUserQuery := `UPDATE users SET is_verified=true WHERE id=$1 RETURNING 
id,email,username,password,name,profile_img,is_verified,role,created_at,updated_at`

	if rollBackErr = tx.QueryRowx(verifyUserQuery, activeUserId).StructScan(&activeUser); rollBackErr != nil {
		return nil, rollBackErr
	}

	// once activated(verified) , remove the user invitation entry for that token

	cleanUpInvitationQuery := `DELETE FROM user_invitations WHERE token=$1`
	_, rollBackErr = tx.Exec(cleanUpInvitationQuery, token)
	if rollBackErr != nil {
		return nil, rollBackErr
	}

	if rollBackErr = tx.Commit(); rollBackErr != nil {
		return nil, rollBackErr
	}

	return &activeUser, nil
}

func (s *Storage) GetUserById(userId int) (*User, error) {

	var user User

	query := `SELECT id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
FROM users WHERE id=$1`

	if err := s.db.QueryRowx(query, userId).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) CreateVerifiedUser(email string, password string) (*User, error) {

	var user User

	query := `INSERT INTO users(email,password,is_verified) VALUES($1,$2,true) RETURNING 
	id,email,username,password,name,profile_img,is_verified,role,created_at,updated_at`

	if err := s.db.QueryRowx(query, email, password).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
