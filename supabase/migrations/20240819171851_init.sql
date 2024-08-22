CREATE TYPE public.game_version AS ENUM ('retail', 'classic');
CREATE TYPE public.addon_provider AS ENUM ('curse', 'github');

CREATE TABLE public.addons (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES auth.users ON DELETE CASCADE,
  game_version game_version NOT NULL,
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  author TEXT NOT NULL,
  url TEXT NOT NULL,
  provider addon_provider NOT NULL,
  provider_id TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, game_version, slug)
);

CREATE INDEX user_id_idx
ON public.addons
USING btree (user_id);

ALTER TABLE public.addons ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can see only their own addons."
ON public.addons FOR SELECT 
USING ( (SELECT auth.uid()) = user_id );

CREATE POLICY "Users can create a addon."
ON public.addons FOR INSERT
TO authenticated                           
WITH CHECK ( (SELECT auth.uid()) = user_id ); 