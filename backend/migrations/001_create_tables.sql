create schema if not exists linguistic;
create schema if not exists learning;

create table linguistic.words (
  id bigserial primary key,
  name text not null,
  normalized_name text not null
);

create table linguistic.concepts (
  id bigserial primary key,
  name text not null,
  description text
);

create table linguistic.gestures (
  id bigserial primary key,
  concept_id bigint not null references linguistic.concepts(id),
  name text not null,
  video_url text,
  description text
);

create table linguistic.translated_words (
  id bigserial primary key,
  word_id bigint not null references linguistic.words(id),
  concept_id bigint not null references linguistic.concepts(id),
  gesture_id bigint references linguistic.gestures(id),
  display_text text not null
);

create table linguistic.lit_examples (
  id bigserial primary key,
  text text not null,
  source text,
  status text not null,
  created_at timestamptz not null default now()
);

create table linguistic.lit_example_segments (
  id bigserial primary key,
  lit_example_id bigint not null references linguistic.lit_examples(id),
  translated_word_id bigint not null references linguistic.translated_words(id),
  text text not null,
  position_index int not null
);

create table learning.roles (
  id bigserial primary key,
  code text not null unique,
  name text not null
);

create table learning.users (
  id bigserial primary key,
  role_id bigint not null references learning.roles(id),
  email text not null unique,
  password_hash text not null,
  full_name text not null,
  created_at timestamptz not null default now()
);

create table learning.courses (
  id bigserial primary key,
  title text not null,
  description text,
  status text not null default 'draft',
  created_by bigint not null references learning.users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table learning.themes (
  id bigserial primary key,
  course_id bigint not null references learning.courses(id),
  title text not null,
  description text,
  order_index int not null,
  status text not null default 'draft',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table learning.theme_translated_words (
  id bigserial primary key,
  theme_id bigint not null references learning.themes(id),
  translated_word_id bigint not null,
  difficulty_level int not null default 1,
  is_required boolean not null default true,
  created_at timestamptz not null default now(),
  unique(theme_id, translated_word_id)
);

create table learning.exercises (
  id bigserial primary key,
  theme_id bigint not null references learning.themes(id),
  lit_example_id bigint not null,
  exercise_type text not null,
  target_mode text not null,
  phrase text not null,
  status text not null default 'draft',
  explanation text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table learning.exercise_segments (
  id bigserial primary key,
  exercise_id bigint not null references learning.exercises(id),
  source_segment_id bigint not null,
  translated_word_id bigint not null,
  word_text text not null,
  gesture_name text not null,
  video_url text,
  answer_visible boolean not null,
  position_index int not null
);

create table learning.generation_runs (
  id bigserial primary key,
  theme_id bigint not null references learning.themes(id),
  started_by bigint not null references learning.users(id),
  status text not null,
  found_examples int not null default 0,
  generated_exercises int not null default 0,
  rejected_examples int not null default 0,
  duration_ms int not null default 0,
  error_message text,
  created_at timestamptz not null default now()
);

create table learning.exercise_reviews (
  id bigserial primary key,
  exercise_id bigint not null references learning.exercises(id),
  reviewer_id bigint not null references learning.users(id),
  decision text not null,
  comment text,
  created_at timestamptz not null default now()
);

create table learning.audit_logs (
  id bigserial primary key,
  user_id bigint references learning.users(id),
  action text not null,
  entity_type text not null,
  entity_id bigint,
  created_at timestamptz not null default now()
);

create index idx_lit_example_segments_example on linguistic.lit_example_segments(lit_example_id);
create index idx_lit_example_segments_tw on linguistic.lit_example_segments(translated_word_id);
create index idx_theme_translated_words_theme on learning.theme_translated_words(theme_id);
create index idx_exercises_theme on learning.exercises(theme_id);
create index idx_exercises_example on learning.exercises(lit_example_id);
create index idx_generation_runs_theme on learning.generation_runs(theme_id);
