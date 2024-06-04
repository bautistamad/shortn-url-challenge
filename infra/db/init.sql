CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE public.url (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shortUrl VARCHAR(255) NOT NULL,
    longUrl TEXT NOT NULL,
    access_count INT DEFAULT 0
);


