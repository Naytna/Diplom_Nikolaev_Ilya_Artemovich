create table if not exists learning.generation_rejections (
    id bigserial primary key,
    generation_run_id bigint not null references learning.generation_runs(id) on delete cascade,
    theme_id bigint not null references learning.themes(id) on delete cascade,
    lit_example_id bigint not null references linguistic.lit_examples(id) on delete cascade,
    example_text text not null,
    reason_code varchar(80) not null,
    reason_text text not null,
    created_at timestamptz not null default now()
);

create index if not exists idx_generation_rejections_run
    on learning.generation_rejections(generation_run_id);

create index if not exists idx_generation_rejections_theme
    on learning.generation_rejections(theme_id);