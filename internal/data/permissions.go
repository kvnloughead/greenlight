package data

import "database/sql"

// Permissions is a string slice for storing permission codes.
type Permissions []string

type Permission struct {
	ID   int64  `json:"id"`
	Code string `json:"string"`
}

type PermissionModel struct {
	DB *sql.DB
}

// Permissions.Includes return a boolean indicating whether a given permission
// code is stored in the calling Permissions slice.
func (p Permissions) Includes(code string) bool {
	for i := range p {
		if p[i] == code {
			return true
		}
	}
	return false
}

// PermissionModel.GetAllForUser retrieves a slice of all permission codes
// associated with the given user ID.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// Join users, permissions, and users_permissions tables to get the permission
	// codes for a given user.
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions 
			ON users_permissions.permission_id = permissions.id
		INNER JOIN users
			ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	// Run the query, inserting the user ID as a placeholder.
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Defer closing the rows.

	// A slice to store the user's permissions.
	var permissions Permissions

	// Iterate through rows, adding each entry to the permissions slice.
	for rows.Next() {
		var code string
		err = rows.Scan(&code)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
