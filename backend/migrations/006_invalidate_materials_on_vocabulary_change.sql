create or replace function learning.invalidate_theme_materials_after_vocabulary_change()
returns trigger as $$
declare
    changed_theme_id bigint;
begin
    changed_theme_id := coalesce(new.theme_id, old.theme_id);

    delete from learning.exercise_reviews
    where exercise_id in (
        select id
        from learning.exercises
        where theme_id = changed_theme_id
    );

    delete from learning.exercise_segments
    where exercise_id in (
        select id
        from learning.exercises
        where theme_id = changed_theme_id
    );

    delete from learning.exercises
    where theme_id = changed_theme_id;

    delete from learning.generation_rejections
    where theme_id = changed_theme_id;

    return coalesce(new, old);
end;
$$ language plpgsql;

drop trigger if exists trg_invalidate_theme_materials_after_vocabulary_insert
on learning.theme_translated_words;

drop trigger if exists trg_invalidate_theme_materials_after_vocabulary_delete
on learning.theme_translated_words;

create trigger trg_invalidate_theme_materials_after_vocabulary_insert
after insert on learning.theme_translated_words
for each row
execute function learning.invalidate_theme_materials_after_vocabulary_change();

create trigger trg_invalidate_theme_materials_after_vocabulary_delete
after delete on learning.theme_translated_words
for each row
execute function learning.invalidate_theme_materials_after_vocabulary_change();