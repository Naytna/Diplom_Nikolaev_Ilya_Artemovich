import { useEffect, useMemo, useState } from 'react'
import './App.css'

type Course = {
  id: number
  title: string
  description: string
  status: string
  themes_count: number | string
  created_at: string
}

type Theme = {
  id: number
  course_id: number
  title: string
  description: string
  order_index: number
  status: string
}

type CourseDetails = {
  course: Course
  themes: Theme[]
}

type VocabularyItem = {
  id: number
  theme_id: number
  translated_word_id: number
  difficulty_level: number
  is_required: boolean
  display_text: string
  word_name: string
  concept_name: string
  concept_description: string
  gesture_name: string
  gesture_description: string
  video_url: string
}

type ExerciseSegment = {
  id: number
  exercise_id: number
  source_segment_id: number
  translated_word_id: number
  word_text: string
  gesture_name: string
  video_url: string
  answer_visible: boolean
  position_index: number
}

type Exercise = {
  id: number
  theme_id: number
  lit_example_id: number
  exercise_type: string
  target_mode: 'textbook' | 'workbook'
  phrase: string
  status: string
  explanation: string
  segments: ExerciseSegment[]
}

type ThemeFull = {
  theme: Theme & {
    course_title: string
  }
  target_mode: 'textbook' | 'workbook'
  vocabulary: VocabularyItem[]
  exercises: Exercise[]
}

const API_URL = 'http://localhost:18080/api'

function App() {
  const [courses, setCourses] = useState<Course[]>([])
  const [courseDetails, setCourseDetails] = useState<CourseDetails | null>(null)
  const [themeFull, setThemeFull] = useState<ThemeFull | null>(null)
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedThemeId, setSelectedThemeId] = useState<number | null>(null)
  const [mode, setMode] = useState<'textbook' | 'workbook'>('textbook')
  const [revealed, setRevealed] = useState<Record<number, boolean>>({})
  const [loading, setLoading] = useState(true)
  const [contentLoading, setContentLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    fetch(`${API_URL}/public/courses`)
      .then((response) => {
        if (!response.ok) {
          throw new Error('Не удалось загрузить список курсов')
        }
        return response.json()
      })
      .then((data: Course[]) => {
        setCourses(data)
        if (data.length > 0) {
          setSelectedCourseId(data[0].id)
        }
      })
      .catch((err: Error) => {
        setError(err.message)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  useEffect(() => {
    if (!selectedCourseId) {
      return
    }

    setContentLoading(true)

    fetch(`${API_URL}/public/courses/${selectedCourseId}`)
      .then((response) => {
        if (!response.ok) {
          throw new Error('Не удалось загрузить курс')
        }
        return response.json()
      })
      .then((data: CourseDetails) => {
        setCourseDetails(data)
        if (data.themes.length > 0) {
          setSelectedThemeId(data.themes[0].id)
        } else {
          setSelectedThemeId(null)
          setThemeFull(null)
        }
      })
      .catch((err: Error) => {
        setError(err.message)
      })
      .finally(() => {
        setContentLoading(false)
      })
  }, [selectedCourseId])

  useEffect(() => {
    if (!selectedThemeId) {
      return
    }

    setContentLoading(true)
    setRevealed({})

    fetch(`${API_URL}/public/themes/${selectedThemeId}/${mode}-full`)
      .then((response) => {
        if (!response.ok) {
          throw new Error('Не удалось загрузить материалы темы')
        }
        return response.json()
      })
      .then((data: ThemeFull) => {
        setThemeFull(data)
      })
      .catch((err: Error) => {
        setError(err.message)
      })
      .finally(() => {
        setContentLoading(false)
      })
  }, [selectedThemeId, mode])

  const currentCourse = useMemo(() => {
    return courses.find((course) => course.id === selectedCourseId) ?? null
  }, [courses, selectedCourseId])

  const toggleAnswer = (exerciseId: number) => {
    setRevealed((prev) => ({
      ...prev,
      [exerciseId]: !prev[exerciseId],
    }))
  }

  if (loading) {
    return <main className="page">Загрузка модуля...</main>
  }

  if (error) {
    return (
      <main className="page">
        <div className="errorBox">{error}</div>
      </main>
    )
  }

  return (
    <main className="page">
      <header className="topbar">
        <div>
          <div className="eyebrow">Модуль генерации учебных заданий</div>
          <h1>Методические материалы РЖЯ</h1>
        </div>
        <div className="statusBadge">public</div>
      </header>

      <section className="workspace">
        <aside className="panel sidebar">
          <div className="panelHeader">
            <h2>Курсы</h2>
          </div>

          {courses.length === 0 && (
            <p className="muted">Опубликованные курсы отсутствуют</p>
          )}

          <div className="list">
            {courses.map((course) => (
              <button
                className={course.id === selectedCourseId ? 'listItem active' : 'listItem'}
                key={course.id}
                onClick={() => setSelectedCourseId(course.id)}
              >
                <span>{course.title}</span>
                <small>Тем: {course.themes_count}</small>
              </button>
            ))}
          </div>
        </aside>

        <aside className="panel sidebar">
          <div className="panelHeader">
            <h2>Темы</h2>
          </div>

          {!currentCourse && (
            <p className="muted">Выберите курс</p>
          )}

          {currentCourse && (
            <div className="courseInfo">
              <strong>{currentCourse.title}</strong>
              <p>{currentCourse.description}</p>
            </div>
          )}

          <div className="list">
            {courseDetails?.themes.map((theme) => (
              <button
                className={theme.id === selectedThemeId ? 'listItem active' : 'listItem'}
                key={theme.id}
                onClick={() => setSelectedThemeId(theme.id)}
              >
                <span>{theme.title}</span>
                <small>№ {theme.order_index}</small>
              </button>
            ))}
          </div>
        </aside>

        <section className="panel content">
          <div className="contentHeader">
            <div>
              <div className="eyebrow">
                {themeFull?.theme.course_title ?? 'Учебный комплект'}
              </div>
              <h2>{themeFull?.theme.title ?? 'Материалы темы'}</h2>
            </div>

            <div className="modeSwitch">
              <button
                className={mode === 'textbook' ? 'modeButton active' : 'modeButton'}
                onClick={() => setMode('textbook')}
              >
                Учебник
              </button>
              <button
                className={mode === 'workbook' ? 'modeButton active' : 'modeButton'}
                onClick={() => setMode('workbook')}
              >
                Рабочая тетрадь
              </button>
            </div>
          </div>

          {contentLoading && (
            <div className="loadingBox">Загрузка материалов...</div>
          )}

          {!contentLoading && themeFull && (
            <>
              <section className="section">
                <div className="sectionHeader">
                  <h3>Лексикон темы</h3>
                  <span>{themeFull.vocabulary.length} элементов</span>
                </div>

                <div className="lexicon">
                  {themeFull.vocabulary.map((item) => (
                    <span key={item.id}>{item.display_text}</span>
                  ))}
                </div>
              </section>

              <section className="section">
                <div className="sectionHeader">
                  <h3>Словарь темы</h3>
                </div>

                <div className="dictionary">
                  {themeFull.vocabulary.map((item) => (
                    <article className="dictCard" key={item.id}>
                      <div>
                        <strong>{item.gesture_name}</strong>
                        <p>{item.gesture_description}</p>
                      </div>
                      <div>
                        <span>{item.display_text}</span>
                        <small>{item.concept_name}</small>
                      </div>
                    </article>
                  ))}
                </div>
              </section>

              <section className="section">
                <div className="sectionHeader">
                  <h3>{mode === 'textbook' ? 'Упражнения учебника' : 'Упражнения рабочей тетради'}</h3>
                  <span>{themeFull.exercises.length} упражнений</span>
                </div>

                {themeFull.exercises.length === 0 && (
                  <p className="muted">Для выбранного режима пока нет опубликованных упражнений</p>
                )}

                <div className="exercises">
                  {themeFull.exercises.map((exercise, index) => {
                    const showAnswer = mode === 'textbook' || revealed[exercise.id]

                    return (
                      <article className="exerciseCard" key={exercise.id}>
                        <div className="exerciseTop">
                          <span>Упражнение {index + 1}</span>
                          <small>{exercise.status}</small>
                        </div>

                        <p className="phrase">{exercise.phrase}</p>

                        {mode === 'workbook' && (
                          <button
                            className="answerButton"
                            onClick={() => toggleAnswer(exercise.id)}
                          >
                            {showAnswer ? 'Скрыть ответ' : 'Показать ответ'}
                          </button>
                        )}

                        {showAnswer && (
                          <div className="segments">
                            {exercise.segments.map((segment) => (
                              <div className="segment" key={segment.id}>
                                <span>{segment.position_index}</span>
                                <strong>{segment.gesture_name}</strong>
                                <small>{segment.word_text}</small>
                              </div>
                            ))}
                          </div>
                        )}
                      </article>
                    )
                  })}
                </div>
              </section>
            </>
          )}
        </section>
      </section>
    </main>
  )
}

export default App