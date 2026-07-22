package campusmap

type Repository interface{ List() ([]Entity, error) }
