-- This migration makes sure content cannot point to a Space owned by another user.
-- It does that with a composite foreign key: (space_id, user_id) must match one row in spaces.

ALTER TABLE spaces
	ADD CONSTRAINT spaces_id_user_id_unique UNIQUE (id, user_id);

ALTER TABLE content
	ADD CONSTRAINT content_space_user_fk
	FOREIGN KEY (space_id, user_id)
	REFERENCES spaces(id, user_id)
	ON DELETE CASCADE;

-- Why this file exists:
-- content already has user_id and space_id.
-- This constraint answers: "Does this content belong to the same user as its Space?"
-- It keeps ownership consistent even if a bug reaches the database layer.
