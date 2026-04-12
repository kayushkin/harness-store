package harnessstore

import (
	"database/sql"
	"time"

	"github.com/kayushkin/llm-bridge/msg"
)

// CreateInstance registers a new harness instance.
func (s *Store) CreateInstance(inst *msg.Instance) error {
	now := time.Now().UTC()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if inst.SSHPort == 0 {
		inst.SSHPort = 22
	}
	_, err := s.db.Exec(`
		INSERT INTO instances (id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inst.ID, inst.HarnessType, inst.Name, inst.Host, inst.Transport,
		inst.SSHUser, inst.SSHKeyPath, inst.SSHPort, inst.WorkingDir,
		inst.MaxConcurrentSessions, inst.Enabled, inst.CreatedAt, inst.UpdatedAt,
	)
	return err
}

// GetInstance retrieves an instance by ID.
func (s *Store) GetInstance(id string) (*msg.Instance, error) {
	var inst msg.Instance
	var enabled int
	err := s.db.QueryRow(`
		SELECT id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at
		FROM instances WHERE id = ?`, id,
	).Scan(&inst.ID, &inst.HarnessType, &inst.Name, &inst.Host, &inst.Transport,
		&inst.SSHUser, &inst.SSHKeyPath, &inst.SSHPort, &inst.WorkingDir,
		&inst.MaxConcurrentSessions, &enabled, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	inst.Enabled = enabled != 0
	return &inst, nil
}

// GetInstanceByName retrieves an instance by name.
func (s *Store) GetInstanceByName(name string) (*msg.Instance, error) {
	var inst msg.Instance
	var enabled int
	err := s.db.QueryRow(`
		SELECT id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at
		FROM instances WHERE name = ?`, name,
	).Scan(&inst.ID, &inst.HarnessType, &inst.Name, &inst.Host, &inst.Transport,
		&inst.SSHUser, &inst.SSHKeyPath, &inst.SSHPort, &inst.WorkingDir,
		&inst.MaxConcurrentSessions, &enabled, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	inst.Enabled = enabled != 0
	return &inst, nil
}

// ListInstances returns all registered instances.
func (s *Store) ListInstances() ([]msg.Instance, error) {
	rows, err := s.db.Query(`
		SELECT id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at
		FROM instances ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []msg.Instance
	for rows.Next() {
		var inst msg.Instance
		var enabled int
		if err := rows.Scan(&inst.ID, &inst.HarnessType, &inst.Name, &inst.Host, &inst.Transport,
			&inst.SSHUser, &inst.SSHKeyPath, &inst.SSHPort, &inst.WorkingDir,
			&inst.MaxConcurrentSessions, &enabled, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
			return nil, err
		}
		inst.Enabled = enabled != 0
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ListInstancesByHarness returns enabled instances for a specific harness type.
func (s *Store) ListInstancesByHarness(harnessType msg.Harness) ([]msg.Instance, error) {
	rows, err := s.db.Query(`
		SELECT id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at
		FROM instances WHERE harness_type = ? AND enabled = 1 ORDER BY name`, harnessType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []msg.Instance
	for rows.Next() {
		var inst msg.Instance
		var enabled int
		if err := rows.Scan(&inst.ID, &inst.HarnessType, &inst.Name, &inst.Host, &inst.Transport,
			&inst.SSHUser, &inst.SSHKeyPath, &inst.SSHPort, &inst.WorkingDir,
			&inst.MaxConcurrentSessions, &enabled, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
			return nil, err
		}
		inst.Enabled = enabled != 0
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ListInstancesByHost returns instances on a specific host.
func (s *Store) ListInstancesByHost(host string) ([]msg.Instance, error) {
	rows, err := s.db.Query(`
		SELECT id, harness_type, name, host, transport, ssh_user, ssh_key_path, ssh_port, working_dir, max_concurrent_sessions, enabled, created_at, updated_at
		FROM instances WHERE host = ? ORDER BY name`, host)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []msg.Instance
	for rows.Next() {
		var inst msg.Instance
		var enabled int
		if err := rows.Scan(&inst.ID, &inst.HarnessType, &inst.Name, &inst.Host, &inst.Transport,
			&inst.SSHUser, &inst.SSHKeyPath, &inst.SSHPort, &inst.WorkingDir,
			&inst.MaxConcurrentSessions, &enabled, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
			return nil, err
		}
		inst.Enabled = enabled != 0
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// UpdateInstance updates an instance's configuration.
func (s *Store) UpdateInstance(inst *msg.Instance) error {
	inst.UpdatedAt = time.Now().UTC()
	res, err := s.db.Exec(`
		UPDATE instances SET harness_type=?, name=?, host=?, transport=?, ssh_user=?, ssh_key_path=?, ssh_port=?, working_dir=?, max_concurrent_sessions=?, enabled=?, updated_at=?
		WHERE id=?`,
		inst.HarnessType, inst.Name, inst.Host, inst.Transport,
		inst.SSHUser, inst.SSHKeyPath, inst.SSHPort, inst.WorkingDir,
		inst.MaxConcurrentSessions, inst.Enabled, inst.UpdatedAt, inst.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteInstance removes an instance and its credential bindings.
func (s *Store) DeleteInstance(id string) error {
	res, err := s.db.Exec(`DELETE FROM instances WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetInstanceEnabled enables or disables an instance.
func (s *Store) SetInstanceEnabled(id string, enabled bool) error {
	res, err := s.db.Exec(`UPDATE instances SET enabled = ?, updated_at = ? WHERE id = ?`,
		enabled, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountInstances returns the total number of instances.
func (s *Store) CountInstances() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM instances`).Scan(&count)
	return count, err
}

// CountEnabledInstances returns the number of enabled instances.
func (s *Store) CountEnabledInstances() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM instances WHERE enabled = 1`).Scan(&count)
	return count, err
}
