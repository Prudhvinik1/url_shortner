CREATE TABLE IF NOT EXISTS public.urls (
  id BIGSERIAL PRIMARY KEY,
  short_code TEXT UNIQUE NOT NULL,
  original_url TEXT NOT NULL,
  is_alias BOOLEAN DEFAULT FALSE,
  ttl BIGINT,
  user_id UUID REFERENCES auth.users(id),
  created_at TIMESTAMPTZ DEFAULT now()
);

