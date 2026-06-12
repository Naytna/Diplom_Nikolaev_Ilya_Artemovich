begin;

-- Удаляем возможные старые дубли упражнений, если они появились до добавления ограничения.
with duplicated_exercises as (
  select id
  from (
    select
      id,
      row_number() over (
        partition by theme_id, lit_example_id, target_mode
        order by id
      ) as rn
    from learning.exercises
  ) ranked
  where rn > 1
)
delete from learning.exercise_reviews
where exercise_id in (select id from duplicated_exercises);

with duplicated_exercises as (
  select id
  from (
    select
      id,
      row_number() over (
        partition by theme_id, lit_example_id, target_mode
        order by id
      ) as rn
    from learning.exercises
  ) ranked
  where rn > 1
)
delete from learning.exercise_segments
where exercise_id in (select id from duplicated_exercises);

with duplicated_exercises as (
  select id
  from (
    select
      id,
      row_number() over (
        partition by theme_id, lit_example_id, target_mode
        order by id
      ) as rn
    from learning.exercises
  ) ranked
  where rn > 1
)
delete from learning.exercises
where id in (select id from duplicated_exercises);

create unique index if not exists idx_exercises_theme_example_mode_unique
on learning.exercises(theme_id, lit_example_id, target_mode);

commit;