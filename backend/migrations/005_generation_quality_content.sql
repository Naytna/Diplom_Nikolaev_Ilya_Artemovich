delete from learning.exercise_reviews
where exercise_id in (
    select id from learning.exercises where theme_id in (1, 101, 102)
);

delete from learning.exercise_segments
where exercise_id in (
    select id from learning.exercises where theme_id in (1, 101, 102)
);

delete from learning.exercises
where theme_id in (1, 101, 102);

delete from learning.generation_rejections
where theme_id in (1, 101, 102);

delete from learning.generation_runs
where theme_id in (1, 101, 102);

delete from learning.theme_translated_words
where theme_id in (101, 102);

insert into linguistic.words (id, name, normalized_name) values
(201, 'я', 'я'),
(202, 'стоять', 'стоять'),
(203, 'лежать', 'лежать'),
(204, 'продавать', 'продавать'),
(205, 'ученик', 'ученик'),
(206, 'дорого', 'дорого'),
(207, 'находиться', 'находиться'),
(208, 'рядом', 'рядом'),
(209, 'быстро', 'быстро'),
(210, 'аккуратно', 'аккуратно'),
(211, 'тетрадь', 'тетрадь'),
(212, 'новый', 'новый'),
(213, 'слева', 'слева')
on conflict (id) do update
set name = excluded.name,
    normalized_name = excluded.normalized_name;

insert into linguistic.concepts (id, name, description) values
(201, 'говорящий человек', 'Понятие первого лица единственного числа'),
(202, 'вертикальное положение', 'Действие или состояние стояния'),
(203, 'горизонтальное положение', 'Действие или состояние лежания'),
(204, 'реализация товара', 'Действие продажи'),
(205, 'обучающийся человек', 'Понятие ученика'),
(206, 'высокая стоимость', 'Признак высокой цены'),
(207, 'расположение в месте', 'Действие нахождения в определённом месте'),
(208, 'близкое расположение', 'Понятие расположения рядом с объектом'),
(209, 'быстрое выполнение действия', 'Признак быстрого выполнения действия'),
(210, 'аккуратное выполнение действия', 'Признак аккуратного выполнения действия'),
(211, 'учебная тетрадь', 'Понятие тетради как учебного предмета'),
(212, 'новый объект', 'Признак нового предмета или объекта'),
(213, 'расположение слева', 'Понятие расположения с левой стороны')
on conflict (id) do update
set name = excluded.name,
    description = excluded.description;

insert into linguistic.gestures (id, concept_id, name, video_url, description) values
(201, 201, 'Я', '/videos/i.mp4', 'Жест для первого лица'),
(202, 202, 'СТОЯТЬ', '/videos/stand.mp4', 'Жест для действия стоять'),
(203, 203, 'ЛЕЖАТЬ', '/videos/lie.mp4', 'Жест для действия лежать'),
(204, 204, 'ПРОДАВАТЬ', '/videos/sell.mp4', 'Жест для действия продавать'),
(205, 205, 'УЧЕНИК', '/videos/student.mp4', 'Жест для ученика'),
(206, 206, 'ДОРОГО', '/videos/expensive.mp4', 'Жест для значения дорого'),
(207, 207, 'НАХОДИТЬСЯ', '/videos/located.mp4', 'Жест для нахождения в месте'),
(208, 208, 'РЯДОМ', '/videos/near.mp4', 'Жест для значения рядом'),
(209, 209, 'БЫСТРО', '/videos/fast.mp4', 'Жест для значения быстро'),
(210, 210, 'АККУРАТНО', '/videos/carefully.mp4', 'Жест для значения аккуратно'),
(211, 211, 'ТЕТРАДЬ', '/videos/notebook.mp4', 'Жест для тетради'),
(212, 212, 'НОВЫЙ', '/videos/new.mp4', 'Жест для значения новый'),
(213, 213, 'СЛЕВА', '/videos/left.mp4', 'Жест для расположения слева')
on conflict (id) do update
set concept_id = excluded.concept_id,
    name = excluded.name,
    video_url = excluded.video_url,
    description = excluded.description;

insert into linguistic.translated_words (id, word_id, concept_id, gesture_id, display_text) values
(201, 201, 201, 201, 'я'),
(202, 202, 202, 202, 'стоять'),
(203, 203, 203, 203, 'лежать'),
(204, 204, 204, 204, 'продавать'),
(205, 205, 205, 205, 'ученик'),
(206, 206, 206, 206, 'дорого'),
(207, 207, 207, 207, 'находиться'),
(208, 208, 208, 208, 'рядом'),
(209, 209, 209, 209, 'быстро'),
(210, 210, 210, 210, 'аккуратно'),
(211, 211, 211, 211, 'тетрадь'),
(212, 212, 212, 212, 'новый/новая/новое'),
(213, 213, 213, 213, 'слева')
on conflict (id) do update
set word_id = excluded.word_id,
    concept_id = excluded.concept_id,
    gesture_id = excluded.gesture_id,
    display_text = excluded.display_text;

update linguistic.lit_examples
set status = 'draft'
where id in (6, 7, 8, 9, 10);

insert into learning.theme_translated_words (theme_id, translated_word_id, difficulty_level, is_required) values
(1, 201, 1, true),
(1, 204, 1, true)
on conflict (theme_id, translated_word_id) do nothing;

insert into learning.theme_translated_words (theme_id, translated_word_id, difficulty_level, is_required) values
(101, 5, 1, true),
(101, 11, 1, true),
(101, 101, 1, true),
(101, 102, 1, true),
(101, 103, 1, true),
(101, 104, 1, true),
(101, 202, 1, true),
(101, 203, 1, true),
(101, 207, 1, true),
(101, 208, 1, true)
on conflict (theme_id, translated_word_id) do nothing;

insert into learning.theme_translated_words (theme_id, translated_word_id, difficulty_level, is_required) values
(102, 201, 1, true),
(102, 105, 1, true),
(102, 106, 1, true),
(102, 107, 1, true),
(102, 108, 1, true),
(102, 109, 1, true),
(102, 110, 1, true),
(102, 205, 1, true),
(102, 211, 1, true)
on conflict (theme_id, translated_word_id) do nothing;

insert into linguistic.lit_examples (id, text, source, status) values
(101, 'В комнате стоят стол и стул', 'Демонстрационный корпус', 'verified'),
(102, 'Комната находится дома', 'Демонстрационный корпус', 'verified'),
(103, 'Дверь находится рядом с окном', 'Демонстрационный корпус', 'verified'),
(104, 'Стул стоит рядом со столом', 'Демонстрационный корпус', 'published'),

(105, 'Я читаю книгу в школе', 'Демонстрационный корпус', 'verified'),
(106, 'Я пишу ручкой в тетради', 'Демонстрационный корпус', 'verified'),
(107, 'Ученик учится читать книгу', 'Демонстрационный корпус', 'published'),
(108, 'Ученик читает книгу в школе', 'Демонстрационный корпус', 'verified'),

(109, 'Абажур стоит рядом со столом', 'Демонстрационный контрольный корпус', 'verified'),
(110, 'Школа находится рядом с домом', 'Демонстрационный контрольный корпус', 'verified'),

(201, 'Я читаю новую книгу', 'Демонстрационный корпус неполной разметки', 'verified'),
(202, 'Ученик пишет аккуратно в тетради', 'Демонстрационный корпус неполной разметки', 'verified'),
(203, 'Дверь находится слева от окна', 'Демонстрационный корпус неполной разметки', 'verified')
on conflict (id) do update
set text = excluded.text,
    source = excluded.source,
    status = excluded.status;

delete from linguistic.lit_example_segments
where lit_example_id in (1, 2, 3, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 201, 202, 203);

insert into linguistic.lit_example_segments (lit_example_id, translated_word_id, text, position_index) values
(1, 201, 'Я', 1),
(1, 4, 'покупаю', 2),
(1, 1, 'хлеб', 3),
(1, 3, 'магазине', 4),

(2, 201, 'Я', 1),
(2, 4, 'покупаю', 2),
(2, 2, 'молоко', 3),
(2, 3, 'магазине', 4),

(3, 3, 'магазине', 1),
(3, 204, 'продают', 2),
(3, 1, 'хлеб', 3),
(3, 2, 'молоко', 4),

(101, 103, 'комнате', 1),
(101, 202, 'стоят', 2),
(101, 101, 'стол', 3),
(101, 102, 'стул', 4),

(102, 103, 'Комната', 1),
(102, 207, 'находится', 2),
(102, 5, 'дома', 3),

(103, 104, 'Дверь', 1),
(103, 207, 'находится', 2),
(103, 208, 'рядом', 3),
(103, 11, 'окном', 4),

(104, 102, 'Стул', 1),
(104, 202, 'стоит', 2),
(104, 208, 'рядом', 3),
(104, 101, 'столом', 4),

(105, 201, 'Я', 1),
(105, 109, 'читаю', 2),
(105, 105, 'книгу', 3),
(105, 106, 'школе', 4),

(106, 201, 'Я', 1),
(106, 108, 'пишу', 2),
(106, 110, 'ручкой', 3),
(106, 211, 'тетради', 4),

(107, 205, 'Ученик', 1),
(107, 107, 'учится', 2),
(107, 109, 'читать', 3),
(107, 105, 'книгу', 4),

(108, 205, 'Ученик', 1),
(108, 109, 'читает', 2),
(108, 105, 'книгу', 3),
(108, 106, 'школе', 4),

(109, 9, 'Абажур', 1),
(109, 202, 'стоит', 2),
(109, 208, 'рядом', 3),
(109, 101, 'столом', 4),

(110, 106, 'Школа', 1),
(110, 207, 'находится', 2),
(110, 208, 'рядом', 3),
(110, 5, 'домом', 4),

(201, 201, 'Я', 1),
(201, 109, 'читаю', 2),
(201, 105, 'книгу', 3),

(202, 205, 'Ученик', 1),
(202, 108, 'пишет', 2),
(202, 211, 'тетради', 3),

(203, 104, 'Дверь', 1),
(203, 207, 'находится', 2),
(203, 11, 'окна', 3);

select setval('linguistic.words_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.words), false);
select setval('linguistic.concepts_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.concepts), false);
select setval('linguistic.gestures_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.gestures), false);
select setval('linguistic.translated_words_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.translated_words), false);
select setval('linguistic.lit_examples_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.lit_examples), false);
select setval('linguistic.lit_example_segments_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.lit_example_segments), false);
select setval('learning.theme_translated_words_id_seq', (select coalesce(max(id), 0) + 1 from learning.theme_translated_words), false);