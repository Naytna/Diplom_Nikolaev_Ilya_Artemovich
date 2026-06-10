import { type FormEvent, useEffect, useState } from 'react'

type Course = {
  id: number
  title: string
  description: string
  status: string
  themes_count?: number | string
}

type Theme = {
  id: number
  course_id: number
  title: string
  description: string
  order_index: number
  status: string
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
  gesture_name: string
}

type TranslatedWord = {
  id: number
  display_text?: string
  word_name?: string
  concept_name?: string
  gesture_name?: string
}

type ExerciseSegment = {
  id: number
  exercise_id: number
  translated_word_id: number
  word_text: string
  gesture_name: string
  position_index: number
}

type Exercise = {
  id: number
  theme_id: number
  exercise_type: string
  target_mode: string
  phrase: string
  status: string
  explanation: string
  segments?: ExerciseSegment[]
}

type GenerationRun = {
  id: number
  run_id?: number
  theme_id: number
  found_examples: number
  generated_exercises: number
  rejected_examples: number
  duration_ms: number
  status: string
  error_message?: string
  created_at?: string
}

type GenerationRejection = {
  id: number
  generation_run_id: number
  theme_id: number
  lit_example_id: number
  example_text: string
  reason_code: string
  reason_text: string
  created_at: string
}

type AuditLog = {
  id: number
  user_id: number
  action: string
  entity_type: string
  entity_id: number
  created_at: string
}

const API_URL = 'http://localhost:18080/api'
type ExpertTab = 'courses' | 'themes' | 'vocabulary' | 'generation' | 'materials' | 'logs'

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers ?? {}),
    },
    ...options,
  })

  const text = await response.text()

  if (!response.ok) {
    throw new Error(text || `Ошибка запроса: ${response.status}`)
  }

  if (!text) {
    return null as T
  }

  return JSON.parse(text) as T
}

function getWordLabel(word: TranslatedWord) {
  return word.display_text || word.word_name || `ID ${word.id}`
}

function getStatusLabel(status: string) {
  if (status === 'completed') return 'завершено'
  if (status === 'failed') return 'ошибка'
  if (status === 'running') return 'выполняется'
  if (status === 'draft') return 'черновик'
  if (status === 'published') return 'опубликовано'
  if (status === 'approved') return 'одобрено'
  if (status === 'rejected') return 'отклонено'

  return status
}

function getModeLabel(mode: string) {
  if (mode === 'textbook') return 'Учебник'
  if (mode === 'workbook') return 'Рабочая тетрадь'

  return mode
}

function getExerciseTypeLabel(type: string) {
  if (type === 'translation_sequence') return 'Последовательность жестов'
  if (type === 'word_matching') return 'Сопоставление слова и жеста'
  if (type === 'sentence_translation') return 'Перевод предложения'

  return type
}

function getActionLabel(action: string) {
  if (action === 'approved') return 'Упражнение одобрено'
  if (action === 'rejected') return 'Упражнение отклонено'
  if (action === 'add_translated_word') return 'Слово добавлено в тему'
  if (action === 'remove_translated_word') return 'Слово удалено из темы'
  if (action === 'create_course') return 'Создан курс'
  if (action === 'create_theme') return 'Создана тема'
  if (action === 'publish_course') return 'Опубликован курс'
  if (action === 'publish_theme') return 'Опубликована тема'
  if (action === 'generate_exercises') return 'Запущена генерация'

  return action
}

function getEntityLabel(entityType: string, entityId: number) {
  if (entityType === 'exercise') return `упражнение №${entityId}`
  if (entityType === 'theme') return `тема №${entityId}`
  if (entityType === 'course') return `курс №${entityId}`
  if (entityType === 'translated_word') return `слово №${entityId}`
  if (entityType === 'generation_run') return `запуск №${entityId}`

  return `${entityType} №${entityId}`
}

function formatDateTime(value?: string) {
  if (!value) {
    return 'дата не указана'
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(date)
}

export default function ExpertPanel() {
  const [activeTab, setActiveTab] = useState<ExpertTab>('courses')
  const [courses, setCourses] = useState<Course[]>([])
  const [themes, setThemes] = useState<Theme[]>([])
  const [themeLabels, setThemeLabels] = useState<Record<number, string>>({})
  const [vocabulary, setVocabulary] = useState<VocabularyItem[]>([])
  const [searchResults, setSearchResults] = useState<TranslatedWord[]>([])
  const [exercises, setExercises] = useState<Exercise[]>([])
  const [runs, setRuns] = useState<GenerationRun[]>([])
  const [audit, setAudit] = useState<AuditLog[]>([])
  const [lastGenerationResult, setLastGenerationResult] = useState<GenerationRun | null>(null)
  const [generationRejections, setGenerationRejections] = useState<GenerationRejection[]>([])
  const [visibleWorkbookAnswers, setVisibleWorkbookAnswers] = useState<Record<number, boolean>>({})

  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedThemeId, setSelectedThemeId] = useState<number | null>(null)

  const [courseTitle, setCourseTitle] = useState('')
  const [courseDescription, setCourseDescription] = useState('')
  const [themeTitle, setThemeTitle] = useState('')
  const [themeDescription, setThemeDescription] = useState('')
  const [themeOrder, setThemeOrder] = useState('1')
  const [search, setSearch] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function loadThemeLabels(courseList: Course[]) {
    const themeGroups = await Promise.all(
      courseList.map(async (course) => {
        try {
          const data = await api<Theme[]>(`/courses/${course.id}/themes`)
          return data ?? []
        } catch {
          return []
        }
      }),
    )

    const labels: Record<number, string> = {}

    themeGroups.flat().forEach((theme) => {
      labels[theme.id] = theme.title
    })

    setThemeLabels(labels)
  }

  async function loadCourses() {
    const data = await api<Course[]>('/courses')
    const safeData = data ?? []

    setCourses(safeData)
    await loadThemeLabels(safeData)

    if (!selectedCourseId && safeData.length > 0) {
      setSelectedCourseId(safeData[0].id)
    }
  }

  async function loadThemes(courseId: number) {
    const data = await api<Theme[]>(`/courses/${courseId}/themes`)
    const safeData = data ?? []

    setThemes(safeData)

    if (safeData.length > 0) {
      setSelectedThemeId(safeData[0].id)
    } else {
      setSelectedThemeId(null)
      setVocabulary([])
      setExercises([])
    }
  }

  async function loadThemeData(themeId: number) {
    const [themeVocabulary, themeExercises] = await Promise.all([
      api<VocabularyItem[]>(`/themes/${themeId}/translated-words`),
      api<Exercise[]>(`/themes/${themeId}/exercises`),
    ])

    setVocabulary(themeVocabulary ?? [])
    setExercises((themeExercises ?? []).map((exercise) => ({
      ...exercise,
      segments: exercise.segments ?? [],
    })))
  }

  const loadGenerationRejections = async (runId: number) => {
    if (!runId) {
      setGenerationRejections([])
      return
    }

    const data = await api<GenerationRejection[]>(`/generation-runs/${runId}/rejections`)
    setGenerationRejections(Array.isArray(data) ? data : [])
  }

  async function loadLogs() {
    const [generationRuns, auditLogs] = await Promise.all([
      api<GenerationRun[]>('/generation-runs'),
      api<AuditLog[]>('/audit'),
    ])

    setRuns(generationRuns ?? [])
    setAudit(auditLogs ?? [])
  }

  async function refreshAll() {
    setError('')
    setMessage('')
    setLoading(true)

    try {
      await loadCourses()
      await loadLogs()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Неизвестная ошибка')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void refreshAll()
  }, [])

  useEffect(() => {
    setLastGenerationResult(null)
    setGenerationRejections([])
    setVisibleWorkbookAnswers({})
  }, [selectedThemeId])


  useEffect(() => {
    if (!selectedCourseId) {
      return
    }

    setError('')
    setLoading(true)

    loadThemes(selectedCourseId)
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Не удалось загрузить темы')
      })
      .finally(() => {
        setLoading(false)
      })
  }, [selectedCourseId])

  useEffect(() => {
    if (!selectedThemeId) {
      return
    }

    setError('')
    setLoading(true)

    loadThemeData(selectedThemeId)
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Не удалось загрузить данные темы')
      })
      .finally(() => {
        setLoading(false)
      })
  }, [selectedThemeId])

  async function createCourse(event: FormEvent) {
    event.preventDefault()

    if (!courseTitle.trim()) {
      setError('Введите название курса')
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      const created = await api<Course>('/courses', {
        method: 'POST',
        body: JSON.stringify({
          title: courseTitle,
          description: courseDescription,
        }),
      })

      setCourseTitle('')
      setCourseDescription('')
      await loadCourses()
      setSelectedCourseId(created.id)
      setMessage('Курс создан')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать курс')
    } finally {
      setLoading(false)
    }
  }

  async function publishCourse() {
    if (!selectedCourseId) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/courses/${selectedCourseId}/publish`, {
        method: 'POST',
      })

      await loadCourses()
      setMessage('Курс опубликован')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось опубликовать курс')
    } finally {
      setLoading(false)
    }
  }

  async function deleteCourse(course: Course) {
    if (course.status !== 'draft') {
      setError('Можно удалить только черновой курс')
      return
    }

    const confirmed = window.confirm(`Удалить курс “${course.title}”? Все темы и материалы курса будут удалены.`)
    if (!confirmed) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/courses/${course.id}`, {
        method: 'DELETE',
      })

      if (selectedCourseId === course.id) {
        setSelectedCourseId(null)
        setSelectedThemeId(null)
        setThemes([])
        setExercises([])
        setGenerationRejections([])
        setLastGenerationResult(null)
      }

      await loadCourses()
      setMessage('Курс удалён')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить курс')
    } finally {
      setLoading(false)
    }
  }

  async function createTheme(event: FormEvent) {
    event.preventDefault()

    if (!selectedCourseId) {
      setError('Сначала выберите курс')
      return
    }

    if (!themeTitle.trim()) {
      setError('Введите название темы')
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      const created = await api<Theme>(`/courses/${selectedCourseId}/themes`, {
        method: 'POST',
        body: JSON.stringify({
          title: themeTitle,
          description: themeDescription,
          order_index: Number(themeOrder) || 1,
        }),
      })

      setThemeTitle('')
      setThemeDescription('')
      setThemeOrder('1')
      await loadThemes(selectedCourseId)
      setSelectedThemeId(created.id)
      setMessage('Тема создана')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать тему')
    } finally {
      setLoading(false)
    }
  }

  async function publishTheme() {
    if (!selectedThemeId) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/themes/${selectedThemeId}/publish`, {
        method: 'POST',
      })

      if (selectedCourseId) {
        await loadThemes(selectedCourseId)
      }

      setMessage('Тема опубликована')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось опубликовать тему')
    } finally {
      setLoading(false)
    }
  }

  async function deleteTheme(theme: Theme) {
    if (theme.status !== 'draft') {
      setError('Можно удалить только черновую тему')
      return
    }

    const confirmed = window.confirm(`Удалить тему “${theme.title}”? Словарь, упражнения и журналы генерации темы будут удалены.`)
    if (!confirmed) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/themes/${theme.id}`, {
        method: 'DELETE',
      })

      if (selectedThemeId === theme.id) {
        setSelectedThemeId(null)
        setExercises([])
        setGenerationRejections([])
        setLastGenerationResult(null)
      }

      if (selectedCourseId) {
        await loadThemes(selectedCourseId)
      }

      setMessage('Тема удалена')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить тему')
    } finally {
      setLoading(false)
    }
  }

  async function searchWords(event: FormEvent) {
    event.preventDefault()

    if (!search.trim()) {
      setSearchResults([])
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      const data = await api<TranslatedWord[]>(`/translated-words?search=${encodeURIComponent(search)}`)
      setSearchResults(data ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось выполнить поиск')
    } finally {
      setLoading(false)
    }
  }

  async function addWordToTheme(translatedWordId: number) {
    if (!selectedThemeId) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/themes/${selectedThemeId}/translated-words`, {
        method: 'POST',
        body: JSON.stringify({
          translated_word_id: translatedWordId,
          difficulty_level: 1,
          is_required: true,
        }),
      })

      await loadThemeData(selectedThemeId)
      setMessage('Слово добавлено в тему')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось добавить слово')
    } finally {
      setLoading(false)
    }
  }

  async function removeWordFromTheme(translatedWordId: number) {
    if (!selectedThemeId) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/themes/${selectedThemeId}/translated-words/${translatedWordId}`, {
        method: 'DELETE',
      })

      await loadThemeData(selectedThemeId)
      setMessage('Слово удалено из темы')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить слово')
    } finally {
      setLoading(false)
    }
  }

  async function generateExercises() {
    if (!selectedThemeId) {
      return
    }

    setError('')
    setMessage('')
    setLoading(true)

    try {
      const result = await api<GenerationRun>(`/themes/${selectedThemeId}/generate`, {
        method: 'POST',
      })

      await loadThemeData(selectedThemeId)
      await loadLogs()

      setLastGenerationResult(result)

      const runId = result.run_id || result.id

      if (runId) {
        await loadGenerationRejections(runId)
      } else {
        setGenerationRejections([])
      }

      setMessage(
        `Генерация завершена: найдено ${result.found_examples}, создано ${result.generated_exercises}, отклонено ${result.rejected_examples}`,
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось запустить генерацию')
    } finally {
      setLoading(false)
    }
  }

  async function approveExercise(exerciseId: number) {
    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/exercises/${exerciseId}/approve`, {
        method: 'PUT',
        body: JSON.stringify({
          reviewer_id: 1,
          comment: 'Проверено через экспертный интерфейс',
        }),
      })

      if (selectedThemeId) {
        await loadThemeData(selectedThemeId)
      }

      setMessage('Упражнение одобрено')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось одобрить упражнение')
    } finally {
      setLoading(false)
    }
  }

  async function rejectExercise(exerciseId: number) {
    setError('')
    setMessage('')
    setLoading(true)

    try {
      await api(`/exercises/${exerciseId}/reject`, {
        method: 'PUT',
        body: JSON.stringify({
          reviewer_id: 1,
          comment: 'Отклонено через экспертный интерфейс',
        }),
      })

      if (selectedThemeId) {
        await loadThemeData(selectedThemeId)
      }

      setMessage('Упражнение отклонено')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось отклонить упражнение')
    } finally {
      setLoading(false)
    }
  }

  const selectedCourse = courses.find((course) => course.id === selectedCourseId)
  const selectedTheme = themes.find((theme) => theme.id === selectedThemeId)

  const exerciseGroups = Object.values(
    exercises.reduce<Record<string, { phrase: string; items: Exercise[] }>>((acc, exercise) => {
      const key = exercise.phrase.trim().toLowerCase()

      if (!acc[key]) {
        acc[key] = {
          phrase: exercise.phrase,
          items: [],
        }
      }

      acc[key].items.push(exercise)

      return acc
    }, {}),
  )

  const previewExercises = exercises
    .filter((exercise) => exercise.status !== 'rejected')
    .sort((a, b) => a.id - b.id)

  const textbookExercises = previewExercises.filter((exercise) => exercise.target_mode === 'textbook')
  const workbookExercises = previewExercises.filter((exercise) => exercise.target_mode === 'workbook')

  const sortedRuns = [...runs].sort((a, b) => b.id - a.id)

  const getThemeLabel = (themeId: number) => {
    if (themeLabels[themeId]) {
      return themeLabels[themeId]
    }

    const theme = themes.find((item) => item.id === themeId)

    if (theme) {
      return theme.title
    }

    return `тема №${themeId}`
  }

  return (
    <section className="expertPage">
      <div className="expertToolbar">
        <div>
          <div className="eyebrow">Экспертная часть</div>
          <h2>Управление генерацией заданий</h2>
          <p>
            Курс: {selectedCourse ? selectedCourse.title : 'не выбран'} · Тема: {selectedTheme ? selectedTheme.title : 'не выбрана'}
          </p>
        </div>

        <button className="primaryButton" onClick={refreshAll}>
          Обновить
        </button>
      </div>

      <nav className="expertTabs">
        <button
          className={activeTab === 'courses' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('courses')}
        >
          Курсы
        </button>
        <button
          className={activeTab === 'themes' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('themes')}
        >
          Темы
        </button>
        <button
          className={activeTab === 'vocabulary' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('vocabulary')}
        >
          Словарь
        </button>
        <button
          className={activeTab === 'generation' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('generation')}
        >
          Генерация
        </button>
        <button
          className={activeTab === 'materials' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('materials')}
        >
          Материалы
        </button>        
        <button
          className={activeTab === 'logs' ? 'expertTab active' : 'expertTab'}
          onClick={() => setActiveTab('logs')}
        >
          Журнал
        </button>
      </nav>



      {(loading || message || error) && (
        <div className="expertNotifications">
          {loading && <div className="loadingBox">Выполняется запрос...</div>}
          {message && <div className="successBox">{message}</div>}
          {error && <div className="errorBox">{error}</div>}
        </div>
      )}

      <div className="expertScreen">
        {activeTab === 'courses' && (
          <section className="expertCard expertSingleScreen">
            <div className="screenHeader">
              <div>
                <h3>Курсы</h3>
                <p>Создание, выбор и публикация учебных курсов</p>
              </div>
            </div>

            <div className="screenGrid">
              <form className="formBlock" onSubmit={createCourse}>
                <input
                  value={courseTitle}
                  onChange={(event) => setCourseTitle(event.target.value)}
                  placeholder="Название курса"
                />
                <textarea
                  value={courseDescription}
                  onChange={(event) => setCourseDescription(event.target.value)}
                  placeholder="Описание курса"
                />
                <button type="submit">Создать курс</button>
              </form>

              <div>
                <div className="adminList">
                  {courses.map((course) => (
                    <div
                      key={course.id}
                      className={course.id === selectedCourseId ? 'adminItemRow active' : 'adminItemRow'}
                    >
                      <button
                        className="adminItemMain"
                        onClick={() => {
                          setSelectedCourseId(course.id)
                          setActiveTab('themes')
                        }}
                      >
                        <strong>{course.title}</strong>
                        <span>{course.description}</span>
                        <small>{getStatusLabel(course.status)}</small>
                      </button>

                      <button
                        className="dangerIconButton"
                        disabled={course.status !== 'draft'}
                        onClick={() => deleteCourse(course)}
                        title={course.status === 'draft' ? 'Удалить курс' : 'Опубликованный курс нельзя удалить'}
                      >
                        Удалить
                      </button>
                    </div>
                  ))}
                </div>

                {selectedCourse && (
                  <button className="secondaryButton fullWidth" onClick={publishCourse}>
                    Опубликовать выбранный курс
                  </button>
                )}
              </div>
            </div>
          </section>
        )}

        {activeTab === 'themes' && (
          <section className="expertCard expertSingleScreen">
            <div className="screenHeader">
              <div>
                <h3>Темы курса</h3>
                <p>{selectedCourse ? `Курс: ${selectedCourse.title}` : 'Сначала выберите курс'}</p>
              </div>
            </div>

            <div className="screenGrid">
              <form className="formBlock" onSubmit={createTheme}>
                <input
                  value={themeTitle}
                  onChange={(event) => setThemeTitle(event.target.value)}
                  placeholder="Название темы"
                />
                <textarea
                  value={themeDescription}
                  onChange={(event) => setThemeDescription(event.target.value)}
                  placeholder="Описание темы"
                />
                <input
                  value={themeOrder}
                  onChange={(event) => setThemeOrder(event.target.value)}
                  placeholder="Порядок"
                />
                <button type="submit">Создать тему</button>
              </form>

              <div>
                <div className="adminList">
                  {themes.map((theme) => (
                    <div
                      key={theme.id}
                      className={theme.id === selectedThemeId ? 'adminItemRow active' : 'adminItemRow'}
                    >
                      <button
                        className="adminItemMain"
                        onClick={() => {
                          setSelectedThemeId(theme.id)
                          setActiveTab('vocabulary')
                        }}
                      >
                        <strong>{theme.title}</strong>
                        <span>{theme.description}</span>
                        <small>Порядок: {theme.order_index} · {getStatusLabel(theme.status)}</small>
                      </button>

                      <button
                        className="dangerIconButton"
                        disabled={theme.status !== 'draft'}
                        onClick={() => deleteTheme(theme)}
                        title={theme.status === 'draft' ? 'Удалить тему' : 'Опубликованную тему нельзя удалить'}
                      >
                        Удалить
                      </button>
                    </div>
                  ))}
                </div>

                {selectedTheme && (
                  <button className="secondaryButton fullWidth" onClick={publishTheme}>
                    Опубликовать выбранную тему
                  </button>
                )}
              </div>
            </div>
          </section>
        )}

        {activeTab === 'vocabulary' && (
          <section className="expertCard expertSingleScreen">
            <div className="screenHeader">
              <div>
                <h3>Словарь выбранной темы</h3>
                <p>
                  {selectedTheme ? `Тема: ${selectedTheme.title} · ${vocabulary.length} слов` : 'Сначала выберите тему'}
                </p>
              </div>
            </div>

            {!selectedTheme && <p className="muted">Выберите тему в разделе “Темы”.</p>}

            {selectedTheme && (
              <>
                <form className="searchBlock" onSubmit={searchWords}>
                  <input
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="Поиск слова, например: хлеб"
                  />
                  <button type="submit">Найти</button>
                </form>

                {searchResults.length > 0 && (
                  <div className="searchResults">
                    {searchResults.map((word) => (
                      <button key={word.id} onClick={() => addWordToTheme(word.id)}>
                        <strong>{getWordLabel(word)}</strong>
                        <span>{word.gesture_name || 'жест'}</span>
                      </button>
                    ))}
                  </div>
                )}

                <div className="lexicon adminLexicon">
                  {vocabulary.map((item) => (
                    <button
                      key={item.id}
                      onClick={() => removeWordFromTheme(item.translated_word_id)}
                      title="Удалить из темы"
                    >
                      {item.display_text || item.word_name}
                    </button>
                  ))}
                </div>
              </>
            )}
          </section>
        )}

        {activeTab === 'generation' && (
          <section className="expertCard expertSingleScreen generationCard">
            <div className="screenHeader">
              <div>
                <h3>Генерация и проверка упражнений</h3>
                <p>
                  {selectedTheme ? `Тема: ${selectedTheme.title}` : 'Сначала выберите тему'}
                </p>
              </div>

              <button className="primaryButton" onClick={generateExercises} disabled={!selectedThemeId}>
                Сгенерировать задания
              </button>
            </div>

            {lastGenerationResult && (
              <div className="generationSummary">
                <div>
                  <span>Найдено примеров</span>
                  <strong>{lastGenerationResult.found_examples}</strong>
                </div>
                <div>
                  <span>Создано упражнений</span>
                  <strong>{lastGenerationResult.generated_exercises}</strong>
                </div>
                <div>
                  <span>Отклонено примеров</span>
                  <strong>{lastGenerationResult.rejected_examples}</strong>
                </div>
                <div>
                  <span>Статус</span>
                  <strong>{getStatusLabel(lastGenerationResult.status)}</strong>
                </div>
              </div>
            )}

            {generationRejections.length > 0 && (
              <section className="rejectionBlock">
                <div className="rejectionHeader">
                  <div>
                    <h4>Отклонённые примеры</h4>
                    <p>Система показывает, какие корпусные примеры не были использованы при генерации и по какой причине.</p>
                  </div>
                  <span>{generationRejections.length}</span>
                </div>

                <div className="rejectionList">
                  {generationRejections.map((item) => (
                    <article className="rejectionItem" key={item.id}>
                      <strong>{item.example_text}</strong>
                      <p>{item.reason_text}</p>
                    </article>
                  ))}
                </div>
              </section>
            )}

            {exerciseGroups.length === 0 && (
              <p className="muted">Для выбранной темы пока нет упражнений</p>
            )}

            <div className="reviewList">
              {exerciseGroups.map((group) => (
                <article className="reviewCard" key={group.phrase}>
                  <div className="reviewTop">
                    <strong>{group.phrase}</strong>
                    <span>{group.items.length} режима</span>
                  </div>

                  <div className="variantList">
                    {group.items.map((exercise) => (
                      <div className="exerciseVariant" key={exercise.id}>
                        <div className="variantTop">
                          <strong>{getModeLabel(exercise.target_mode)}</strong>
                          <span>{getStatusLabel(exercise.status)}</span>
                        </div>

                        <p>Тип: {getExerciseTypeLabel(exercise.exercise_type)}</p>

                        {exercise.segments && exercise.segments.length > 0 && (
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

                        <div className="reviewActions">
                          <button onClick={() => approveExercise(exercise.id)}>
                            Одобрить
                          </button>
                          <button onClick={() => rejectExercise(exercise.id)}>
                            Отклонить
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                </article>
              ))}
            </div>
          </section>
        )}

        {activeTab === 'materials' && (
          <section className="expertCard expertSingleScreen materialsScreen">
            <div className="screenHeader">
              <div>
                <h3>Предпросмотр учебных материалов</h3>
                <p>
                  {selectedTheme
                    ? `Тема: ${selectedTheme.title}. Эксперт видит, как материалы будут выглядеть для ученика.`
                    : 'Сначала выберите тему'}
                </p>
              </div>

              {selectedTheme && (
                <button className="primaryButton" onClick={publishTheme}>
                  Опубликовать тему
                </button>
              )}
            </div>

            {!selectedTheme && <p className="muted">Выберите тему в разделе “Темы”.</p>}

            {selectedTheme && textbookExercises.length === 0 && workbookExercises.length === 0 && (
              <p className="muted">
                Для выбранной темы пока нет сгенерированных материалов. Сначала запустите генерацию заданий.
              </p>
            )}

            {selectedTheme && (textbookExercises.length > 0 || workbookExercises.length > 0) && (
              <div className="materialsPreviewGrid">
                <article className="materialSheet">
                  <div className="materialSheetHeader">
                    <div>
                      <span>Учебник</span>
                      <h4>{selectedTheme.title}</h4>
                    </div>
                    <small>{textbookExercises.length} заданий</small>
                  </div>

                  <div className="materialList">
                    {textbookExercises.map((exercise, index) => (
                      <section className="materialExercise" key={exercise.id}>
                        <div className="previewMeta">
                          <span>Задание {index + 1}</span>
                          <strong>{getStatusLabel(exercise.status)}</strong>
                        </div>

                        <p className="materialQuestion">{exercise.phrase}</p>

                        <div className="materialAnswer">
                          <span>Последовательность жестов:</span>
                          <div className="answerSequence">
                            {(exercise.segments ?? []).map((segment) => (
                              <strong key={segment.id}>{segment.gesture_name}</strong>
                            ))}
                          </div>
                        </div>

                        {exercise.explanation && (
                          <p className="materialExplanation">{exercise.explanation}</p>
                        )}
                      </section>
                    ))}
                  </div>
                </article>

                <article className="materialSheet workbookSheet">
                  <div className="materialSheetHeader">
                    <div>
                      <span>Рабочая тетрадь</span>
                      <h4>{selectedTheme.title}</h4>
                    </div>
                    <small>{workbookExercises.length} заданий</small>
                  </div>

                  <div className="materialList">
                    {workbookExercises.map((exercise, index) => {
                      const isAnswerVisible = Boolean(visibleWorkbookAnswers[exercise.id])
                      const segments = exercise.segments ?? []

                      return (
                        <section className="materialExercise" key={exercise.id}>
                          <div className="previewMeta">
                            <span>Задание {index + 1}</span>
                            <strong>{getStatusLabel(exercise.status)}</strong>
                          </div>

                          <p className="materialQuestion">{exercise.phrase}</p>

                          <div className="hiddenAnswer">
                            <span>Запишите последовательность жестов:</span>

                            {!isAnswerVisible && (
                              <div className="workbookSlots">
                                {segments.map((segment) => (
                                  <span className="answerSlot" key={segment.id}>
                                    {segment.position_index}
                                  </span>
                                ))}
                              </div>
                            )}

                            {isAnswerVisible && (
                              <div className="answerSequence">
                                {segments.map((segment) => (
                                  <strong key={segment.id}>{segment.gesture_name}</strong>
                                ))}
                              </div>
                            )}
                          </div>

                          <button
                            className="answerButton"
                            onClick={() => {
                              setVisibleWorkbookAnswers((current) => ({
                                ...current,
                                [exercise.id]: !current[exercise.id],
                              }))
                            }}
                          >
                            {isAnswerVisible ? 'Скрыть ответ' : 'Показать ответ'}
                          </button>
                        </section>
                      )
                    })}
                  </div>
                </article>
              </div>
            )}
          </section>
        )}

        {activeTab === 'logs' && (
          <div className="logsScreen">
            <section className="expertCard">
              <h3>Журнал генераций</h3>

              <div className="compactLog">
                {sortedRuns.slice(0, 10).map((run) => (
                  <div key={run.id}>
                    <strong>Запуск #{run.id}</strong>
                    <span>Тема: {getThemeLabel(run.theme_id)}</span>
                    <span>{formatDateTime(run.created_at)}</span>
                    <span>
                      {getStatusLabel(run.status)}: найдено {run.found_examples}, создано {run.generated_exercises}, отклонено {run.rejected_examples}
                    </span>
                  </div>
                ))}
              </div>
            </section>

            <section className="expertCard">
              <h3>История действий эксперта</h3>

              <div className="compactLog">
                {audit.slice(0, 10).map((item) => (
                  <div key={item.id}>
                    <strong>{getActionLabel(item.action)}</strong>
                    <span>{getEntityLabel(item.entity_type, item.entity_id)}</span>
                  </div>
                ))}
              </div>
            </section>
          </div>
        )}
      </div>
    </section>
  )
}