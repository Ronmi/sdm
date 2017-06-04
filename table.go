package sdm

// CreateTables creates all known table, breaks at first error
func (m *Manager) CreateTables() (err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for t, n := range m.table {
		_, err = m.drv.CreateTable(
			m.Connection(),
			n,
			t,
			m.fields[t],
			m.indexes[t],
		)
		if err != nil {
			return
		}
	}

	return
}

// CreateTablesNotExist creates all known table only if table yet created
func (m *Manager) CreateTablesNotExist() (err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for t, n := range m.table {
		_, err = m.drv.CreateTableNotExist(
			m.Connection(),
			n,
			t,
			m.fields[t],
			m.indexes[t],
		)
		if err != nil {
			return
		}
	}

	return
}
