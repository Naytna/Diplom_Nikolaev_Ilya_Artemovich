insert into linguistic.words (id, name, normalized_name) values
(101, 'стол', 'стол'),
(102, 'стул', 'стул'),
(103, 'комната', 'комната'),
(104, 'дверь', 'дверь'),
(105, 'книга', 'книга'),
(106, 'школа', 'школа'),
(107, 'учиться', 'учиться'),
(108, 'писать', 'писать'),
(109, 'читать', 'читать'),
(110, 'ручка', 'ручка')
on conflict (id) do nothing;

insert into linguistic.concepts (id, name, description) values
(101, 'предмет мебели для работы', 'Понятие стола как предмета мебели'),
(102, 'предмет мебели для сидения', 'Понятие стула'),
(103, 'часть жилого помещения', 'Понятие комнаты'),
(104, 'элемент входа в помещение', 'Понятие двери'),
(105, 'печатное издание', 'Понятие книги'),
(106, 'учебное заведение', 'Понятие школы'),
(107, 'получение знаний', 'Действие обучения'),
(108, 'письменное действие', 'Действие письма'),
(109, 'восприятие текста', 'Действие чтения'),
(110, 'предмет для письма', 'Понятие ручки')
on conflict (id) do nothing;

insert into linguistic.gestures (id, concept_id, name, video_url, description) values
(101, 101, 'СТОЛ', '/videos/table.mp4', 'Жест для стола'),
(102, 102, 'СТУЛ', '/videos/chair.mp4', 'Жест для стула'),
(103, 103, 'КОМНАТА', '/videos/room.mp4', 'Жест для комнаты'),
(104, 104, 'ДВЕРЬ', '/videos/door.mp4', 'Жест для двери'),
(105, 105, 'КНИГА', '/videos/book.mp4', 'Жест для книги'),
(106, 106, 'ШКОЛА', '/videos/school.mp4', 'Жест для школы'),
(107, 107, 'УЧИТЬСЯ', '/videos/study.mp4', 'Жест для обучения'),
(108, 108, 'ПИСАТЬ', '/videos/write.mp4', 'Жест для письма'),
(109, 109, 'ЧИТАТЬ', '/videos/read.mp4', 'Жест для чтения'),
(110, 110, 'РУЧКА', '/videos/pen.mp4', 'Жест для ручки')
on conflict (id) do nothing;

insert into linguistic.translated_words (id, word_id, concept_id, gesture_id, display_text) values
(101, 101, 101, 101, 'стол'),
(102, 102, 102, 102, 'стул'),
(103, 103, 103, 103, 'комната'),
(104, 104, 104, 104, 'дверь'),
(105, 105, 105, 105, 'книга'),
(106, 106, 106, 106, 'школа'),
(107, 107, 107, 107, 'учиться'),
(108, 108, 108, 108, 'писать'),
(109, 109, 109, 109, 'читать'),
(110, 110, 110, 110, 'ручка')
on conflict (id) do nothing;

insert into learning.themes (id, course_id, title, description, order_index, status) values
(101, 1, 'Дом и быт', 'Лексика для описания дома, комнаты и бытовых предметов', 3, 'draft'),
(102, 1, 'Учёба', 'Лексика для учебных действий и предметов', 4, 'draft')
on conflict (id) do nothing;

insert into learning.theme_translated_words (theme_id, translated_word_id, difficulty_level, is_required) values
(101, 5, 1, true),
(101, 11, 1, true),
(101, 13, 1, true),
(101, 101, 1, true),
(101, 102, 1, true),
(101, 103, 1, true),
(101, 104, 1, true),
(102, 101, 1, true),
(102, 105, 1, true),
(102, 106, 1, true),
(102, 107, 1, true),
(102, 108, 1, true),
(102, 109, 1, true),
(102, 110, 1, true)
on conflict (theme_id, translated_word_id) do nothing;

insert into linguistic.lit_examples (id, text, source, status) values
(101, 'В комнате есть стол и стул', 'Демонстрационный корпус', 'verified'),
(102, 'Моя комната дома', 'Демонстрационный корпус', 'verified'),
(103, 'Дверь и окно дома', 'Демонстрационный корпус', 'verified'),
(104, 'Стул стоит у стола', 'Демонстрационный корпус', 'published'),
(105, 'Я читаю книгу в школе', 'Демонстрационный корпус', 'verified'),
(106, 'Я пишу ручкой в школе', 'Демонстрационный корпус', 'verified'),
(107, 'Ученик учится и читает книгу', 'Демонстрационный корпус', 'published'),
(108, 'Книга лежит на столе', 'Демонстрационный корпус', 'verified'),
(109, 'Абажур стоит на столе', 'Демонстрационный корпус', 'verified'),
(110, 'Школа находится дома', 'Демонстрационный корпус', 'verified')
on conflict (id) do nothing;

insert into linguistic.lit_example_segments (lit_example_id, translated_word_id, text, position_index) values
(101, 103, 'комнате', 1),
(101, 101, 'стол', 2),
(101, 102, 'стул', 3),
(102, 13, 'Моя', 1),
(102, 103, 'комната', 2),
(102, 5, 'дома', 3),
(103, 104, 'Дверь', 1),
(103, 11, 'окно', 2),
(103, 5, 'дома', 3),
(104, 102, 'Стул', 1),
(104, 101, 'стола', 2),
(105, 109, 'читаю', 1),
(105, 105, 'книгу', 2),
(105, 106, 'школе', 3),
(106, 108, 'пишу', 1),
(106, 110, 'ручкой', 2),
(106, 106, 'школе', 3),
(107, 107, 'учится', 1),
(107, 109, 'читает', 2),
(107, 105, 'книгу', 3),
(108, 105, 'Книга', 1),
(108, 101, 'столе', 2),
(109, 9, 'Абажур', 1),
(109, 101, 'столе', 2),
(110, 106, 'Школа', 1),
(110, 5, 'дома', 2)
on conflict do nothing;

select setval('linguistic.words_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.words), false);
select setval('linguistic.concepts_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.concepts), false);
select setval('linguistic.gestures_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.gestures), false);
select setval('linguistic.translated_words_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.translated_words), false);
select setval('linguistic.lit_examples_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.lit_examples), false);
select setval('linguistic.lit_example_segments_id_seq', (select coalesce(max(id), 0) + 1 from linguistic.lit_example_segments), false);
select setval('learning.themes_id_seq', (select coalesce(max(id), 0) + 1 from learning.themes), false);
select setval('learning.theme_translated_words_id_seq', (select coalesce(max(id), 0) + 1 from learning.theme_translated_words), false);