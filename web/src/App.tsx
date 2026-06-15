import { useCallback, useEffect, useMemo, useState } from 'react'
import './App.css'
import ExpertPanel from './ExpertPanel'
import { api, publicApi, setDemoRole, type AuthUser, type DemoRole } from './api'

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

type AuthState = {
  token: string | null
  user: AuthUser | null
}

type LoginResponse = {
  token: string
  user: AuthUser
}

type ViewMode = 'public' | 'expert' | 'login'

const STORAGE_KEY = 'rsl-demo-auth'

type RolePermissions = {
  canViewExpertPanel: boolean
  canViewWorkbook: boolean
  canUseSelfCheck: boolean
  canPublish: boolean
  canGenerate: boolean
}

const rolePermissions: Record<DemoRole, RolePermissions> = {
  guest: {
    canViewExpertPanel: false,
    canViewWorkbook: false,
    canUseSelfCheck: false,
    canPublish: false,
    canGenerate: false,
  },
  learner: {
    canViewExpertPanel: false,
    canViewWorkbook: true,
    canUseSelfCheck: true,
    canPublish: false,
    canGenerate: false,
  },
  expert: {
    canViewExpertPanel: true,
    canViewWorkbook: true,
    canUseSelfCheck: true,
    canPublish: true,
    canGenerate: true,
  },
}

function normalizeThemeFull(data: ThemeFull): ThemeFull {
  return {
    ...data,
    vocabulary: data.vocabulary ?? [],
    exercises: (data.exercises ?? []).map((exercise) => ({
      ...exercise,
      segments: exercise.segments ?? [],
    })),
  }
}

function isExpert(user: AuthUser | null) {
  return user?.role === 'expert'
}

function saveAuthState(state: AuthState) {
  if (!state.token || !state.user) {
    localStorage.removeItem(STORAGE_KEY)
    return
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
}

function loadStoredAuthState(): AuthState {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return { token: null, user: null }
  }

  try {
    const parsed = JSON.parse(raw) as AuthState
    if (!parsed.token || !parsed.user) {
      return { token: null, user: null }
    }

    return parsed
  } catch {
    return { token: null, user: null }
  }
}

function LoginScreen({
  loading,
  error,
  onBack,
  onSubmit,
}: {
  loading: boolean
  error: string
  onBack: () => void
  onSubmit: (username: string, password: string) => Promise<void>
}) {
  const [username, setUsername] = useState('expert')
  const [password, setPassword] = useState('expert123')

  const applyDemoCredentials = (nextUsername: string, nextPassword: string) => {
    setUsername(nextUsername)
    setPassword(nextPassword)
  }

  return (
    <section className="loginScreen">
      <div className="loginPanel">
        <div className="eyebrow">Демонстрационный доступ эксперта</div>
        <h2>Вход в модуль</h2>
        <p className="loginLead">
          Эксперт работает с генерацией, публикацией и журналом. Обучающийся изучает опубликованные материалы и работает с рабочей тетрадью.
        </p>

        <div className="demoAccounts">
          <button
            className="demoAccountCard"
            type="button"
            onClick={() => applyDemoCredentials('expert', 'expert123')}
          >
            <strong>Эксперт</strong>
            <span>expert / expert123</span>
          </button>
          <button
            className="demoAccountCard"
            type="button"
            onClick={() => applyDemoCredentials('student', 'student123')}
          >
            <strong>Обучающийся</strong>
            <span>student / student123</span>
          </button>
        </div>

        <form
          className="loginForm"
          onSubmit={async (event) => {
            event.preventDefault()
            await onSubmit(username, password)
          }}
        >
          <label>
            <span>Username</span>
            <input
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="expert или student"
            />
          </label>

          <label>
            <span>Password</span>
            <input
              autoComplete="current-password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Введите пароль"
            />
          </label>

          {error && <div className="errorBox">{error}</div>}

          <div className="loginActions">
            <button className="secondaryButton" type="button" onClick={onBack}>
              Вернуться к публичной части
            </button>
            <button className="primaryButton" type="submit" disabled={loading}>
              {loading ? 'Выполняется вход...' : 'Войти'}
            </button>
          </div>
        </form>
      </div>
    </section>
  )
}

function App() {
  const [viewMode, setViewMode] = useState<ViewMode>('public')
  const [authState, setAuthState] = useState<AuthState>({ token: null, user: null })
  const [authLoading, setAuthLoading] = useState(true)
  const [loginLoading, setLoginLoading] = useState(false)
  const [loginError, setLoginError] = useState('')
  const [courses, setCourses] = useState<Course[]>([])
  const [courseDetails, setCourseDetails] = useState<CourseDetails | null>(null)
  const [themeFull, setThemeFull] = useState<ThemeFull | null>(null)
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedThemeId, setSelectedThemeId] = useState<number | null>(null)
  const [mode, setMode] = useState<'textbook' | 'workbook'>('textbook')
  const [revealed, setRevealed] = useState<Record<number, boolean>>({})
  const [workbookAnswers, setWorkbookAnswers] = useState<Record<number, string>>({})
  const [loading, setLoading] = useState(true)
  const [contentLoading, setContentLoading] = useState(false)
  const [error, setError] = useState('')
  const [roleMessage, setRoleMessage] = useState('')
  const [publicRefreshToken, setPublicRefreshToken] = useState(0)

  const demoRole: DemoRole =
    authState.user?.role === 'expert'
      ? 'expert'
      : authState.user?.role === 'student'
        ? 'learner'
        : 'guest'
  const permissions = rolePermissions[demoRole]

  const clearAuth = useCallback(() => {
    const nextState = { token: null, user: null }
    setAuthState(nextState)
    saveAuthState(nextState)
    setDemoRole('guest')
    setMode('textbook')
    setRoleMessage('')
    setViewMode('public')
  }, [])

  const applyAuth = useCallback((token: string, user: AuthUser) => {
    const nextState = { token, user }
    setAuthState(nextState)
    saveAuthState(nextState)
    setDemoRole(user.role === 'expert' ? 'expert' : 'learner')
    setLoginError('')
    setViewMode(user.role === 'expert' ? 'expert' : 'public')
  }, [])

  useEffect(() => {
    const stored = loadStoredAuthState()
    setDemoRole('guest')

    if (!stored.token) {
      setAuthLoading(false)
      return
    }

    api<AuthUser>('/auth/me', {
      authToken: stored.token,
      onUnauthorized: clearAuth,
    })
      .then((user) => {
        applyAuth(stored.token!, user)
      })
      .catch(() => {
        clearAuth()
      })
      .finally(() => {
        setAuthLoading(false)
      })
  }, [applyAuth, clearAuth])

  const loadPublicCourses = useCallback((silent = false) => {
    if (!silent) {
      setLoading(true)
    }

    return publicApi<Course[]>('/public/courses')
      .then((data: Course[] | null) => {
        const safeData = Array.isArray(data) ? data : []

        setCourses(safeData)
        setError('')

        if (safeData.length > 0) {
          setSelectedCourseId((current) => {
            if (current && safeData.some((course) => course.id === current)) {
              return current
            }

            return safeData[0].id
          })
        } else {
          setSelectedCourseId(null)
          setCourseDetails(null)
          setSelectedThemeId(null)
          setThemeFull(null)
        }
      })
      .catch((err: Error) => {
        setError(err.message)
      })
      .finally(() => {
        if (!silent) {
          setLoading(false)
        }
      })
  }, [])

  useEffect(() => {
    loadPublicCourses()
  }, [loadPublicCourses])

  useEffect(() => {
    if (!permissions.canViewWorkbook && mode === 'workbook') {
      setRoleMessage('Рабочая тетрадь доступна только обучающемуся')
      setMode('textbook')
    }
  }, [mode, permissions.canViewWorkbook])

  useEffect(() => {
    if (!selectedCourseId) {
      return
    }

    setContentLoading(true)

    publicApi<CourseDetails>(`/public/courses/${selectedCourseId}`)
      .then((data: CourseDetails | null) => {
        if (!data) {
          setCourseDetails(null)
          setSelectedThemeId(null)
          setThemeFull(null)
          return
        }

        const safeData = {
          ...data,
          themes: Array.isArray(data.themes) ? data.themes : [],
        }

        setCourseDetails(safeData)
        setError('')

        if (safeData.themes.length > 0) {
          setSelectedThemeId((current) => {
            if (current && safeData.themes.some((theme) => theme.id === current)) {
              return current
            }

            return safeData.themes[0].id
          })
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
  }, [selectedCourseId, publicRefreshToken])

  useEffect(() => {
    if (!selectedThemeId) {
      setThemeFull(null)
      return
    }

    if (courseDetails && !courseDetails.themes.some((theme) => theme.id === selectedThemeId)) {
      setThemeFull(null)
      return
    }

    setContentLoading(true)
    setRevealed({})
    setWorkbookAnswers({})

    publicApi<ThemeFull>(`/public/themes/${selectedThemeId}/${mode}-full`)
      .then((data: ThemeFull | null) => {
        if (!data) {
          setThemeFull(null)
          return
        }

        setThemeFull(normalizeThemeFull(data))
        setError('')
      })
      .catch((err: Error) => {
        if (err.message === 'опубликованная тема не найдена') {
          setThemeFull(null)
          return
        }

        setError(err.message)
      })
      .finally(() => {
        setContentLoading(false)
      })
  }, [selectedThemeId, mode, courseDetails, publicRefreshToken])

  const safeCourses = useMemo(() => {
    return Array.isArray(courses) ? courses : []
  }, [courses])

  const currentCourse = useMemo(() => {
    return safeCourses.find((course) => course.id === selectedCourseId) ?? null
  }, [safeCourses, selectedCourseId])

  const toggleAnswer = (exerciseId: number) => {
    setRevealed((prev) => ({
      ...prev,
      [exerciseId]: !prev[exerciseId],
    }))
  }

  const updateWorkbookAnswer = (exerciseId: number, value: string) => {
    setWorkbookAnswers((prev) => ({
      ...prev,
      [exerciseId]: value,
    }))
  }

  const handleLogin = async (username: string, password: string) => {
    setLoginLoading(true)
    setLoginError('')

    try {
      const result = await api<LoginResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      })

      applyAuth(result.token, result.user)
    } catch (err) {
      setLoginError(err instanceof Error ? err.message : 'Не удалось выполнить вход')
    } finally {
      setLoginLoading(false)
    }
  }

  const openExpertArea = () => {
    if (!permissions.canViewExpertPanel) {
      setRoleMessage('Экспертная часть доступна только пользователю с ролью эксперта')
      setViewMode('public')
      return
    }

    setRoleMessage('')

    if (!authState.user || authState.user.role !== 'expert' || !authState.token) {
      setViewMode('login')
      return
    }

    setViewMode('expert')
  }

  const openWorkbook = () => {
    if (!permissions.canViewWorkbook) {
      setRoleMessage('Рабочая тетрадь доступна только обучающемуся')
      setMode('textbook')
      return
    }

    setRoleMessage('')
    setMode('workbook')
  }

  if (authLoading || loading) {
    return <main className="page">Загрузка модуля...</main>
  }

  if (viewMode === 'login') {
    return (
      <main className="page">
        <header className="topbar">
          <div className="brandBlock">
            <div className="eyebrow">Модуль генерации учебных заданий</div>
            <h1 className="topbarTitle">Методические материалы РЖЯ</h1>
          </div>
        </header>
        <LoginScreen
          error={loginError}
          loading={loginLoading}
          onBack={() => setViewMode('public')}
          onSubmit={handleLogin}
        />
      </main>
    )
  }

  return (
    <main className="page">
      <header className="topbar">
        <div className="brandBlock">
          <div className="eyebrow">Модуль генерации учебных заданий</div>
          <h1 className="topbarTitle">Методические материалы РЖЯ</h1>
        </div>
        <div className="topActions">
          <button
            className={viewMode === 'public' ? 'navButton active' : 'navButton'}
            onClick={() => setViewMode('public')}
          >
            Публичная часть
          </button>

          {isExpert(authState.user) && (
            <button
              className={viewMode === 'expert' ? 'navButton active' : 'navButton'}
              onClick={openExpertArea}
            >
              Экспертная часть
            </button>
          )}

          {!authState.user && (
            <button className="navButton" onClick={() => setViewMode('login')}>
              Войти
            </button>
          )}
        </div>

        <div className="sessionPanel">
          {authState.user ? (
            <>
              <div className="sessionBadge">
                <strong>{authState.user.full_name}</strong>
                <span className="sessionRole">
                  {authState.user.role === 'expert' ? 'Роль: эксперт' : 'Роль: обучающийся'}
                </span>
              </div>
              <button className="secondaryButton" onClick={clearAuth}>
                Выйти
              </button>
            </>
          ) : (
            <div className="sessionBadge guest">
              <strong>Гость</strong>
              <span>Доступны курсы, темы, словарь и учебник</span>
            </div>
          )}
        </div>
      </header>

      {roleMessage && <div className="roleNotice">{roleMessage}</div>}

      {viewMode === 'expert' && isExpert(authState.user) && authState.token ? (
        <ExpertPanel
          authToken={authState.token}
          currentUser={authState.user as AuthUser}
          onUnauthorized={clearAuth}
          onPublicContentChanged={async () => {
            await loadPublicCourses(true)
            setPublicRefreshToken((current) => current + 1)
          }}
        />
      ) : (
        <section className="workspace">
          <aside className="panel sidebar">
            <div className="panelHeader">
              <h2>Курсы</h2>
            </div>

            {safeCourses.length === 0 && (
              <p className="muted">Опубликованные курсы отсутствуют</p>
            )}

            <div className="list">
              {safeCourses.map((course) => (
                <button
                  className={course.id === selectedCourseId ? 'listItem active' : 'listItem'}
                  key={course.id}
                  onClick={() => setSelectedCourseId(course.id)}
                >
                  <div className="listItemBody">
                    <span>{course.title}</span>
                    <small>{course.description}</small>
                  </div>
                  <div className="listItemMeta">
                    <small>Тем: {course.themes_count}</small>
                  </div>
                </button>
              ))}
            </div>
          </aside>

          <aside className="panel sidebar">
            <div className="panelHeader">
              <h2>Темы</h2>
            </div>

            {currentCourse ? (
              <div className="courseContextCard">
                <strong>{currentCourse.title}</strong>
                <p>{currentCourse.description}</p>
                <div className="contextMeta">
                  <span>Опубликованных тем: {currentCourse.themes_count}</span>
                </div>
              </div>
            ) : (
              <p className="muted">Выберите курс для просмотра тем</p>
            )}

            <div className="listSectionHeader">
              <span>Темы курса</span>
              <small>{courseDetails?.themes?.length ?? 0}</small>
            </div>

            {selectedCourseId && courseDetails && (courseDetails.themes ?? []).length === 0 && (
              <p className="muted">У выбранного курса пока нет опубликованных тем</p>
            )}

            <div className="themeList">
              {(courseDetails?.themes ?? []).map((theme) => (
                <button
                  className={theme.id === selectedThemeId ? 'themeListItem active' : 'themeListItem'}
                  key={theme.id}
                  onClick={() => setSelectedThemeId(theme.id)}
                >
                  <div>
                    <span>{theme.title}</span>
                    <small>{theme.description}</small>
                  </div>
                  <em>№ {theme.order_index}</em>
                </button>
              ))}
            </div>
          </aside>

          <section className="panel content">
            <div className="contentHeader">
              <div className="contentHeading">
                <div className="eyebrow">
                  {themeFull?.theme.course_title ?? 'Учебный комплект'}
                </div>
                <h2>{themeFull?.theme.title ?? 'Материалы темы'}</h2>
                {themeFull?.theme.description && (
                  <p className="contentLead">{themeFull.theme.description}</p>
                )}
              </div>

              <div className="modeSwitch">
                <button
                  className={mode === 'textbook' ? 'modeButton active' : 'modeButton'}
                  onClick={() => {
                    setRoleMessage('')
                    setMode('textbook')
                  }}
                >
                  Учебник
                </button>
                {permissions.canViewWorkbook && (
                  <button
                    className={mode === 'workbook' ? 'modeButton active' : 'modeButton'}
                    onClick={openWorkbook}
                  >
                    Рабочая тетрадь
                  </button>
                )}
              </div>
            </div>

            {error && <div className="errorBox publicError">{error}</div>}

            {contentLoading && (
              <div className="loadingBox">Загрузка материалов...</div>
            )}

            {!contentLoading && !themeFull && (
              <div className="emptyState">
                {!selectedCourseId ? (
                  <>
                    <h3>Выберите курс</h3>
                    <p>Для просмотра материалов выберите опубликованный курс в левом списке.</p>
                  </>
                ) : courseDetails && (courseDetails.themes ?? []).length === 0 ? (
                  <>
                    <h3>В курсе пока нет опубликованных тем</h3>
                    <p>
                      Материалы появятся после публикации темы и одобрения упражнений в экспертной части.
                    </p>
                  </>
                ) : !selectedThemeId ? (
                  <>
                    <h3>Выберите тему</h3>
                    <p>Для просмотра учебника или рабочей тетради выберите тему курса.</p>
                  </>
                ) : (
                  <>
                    <h3>Материалы темы пока недоступны</h3>
                    <p>
                      Для выбранной темы пока нет опубликованных материалов в режиме учебника или рабочей тетради.
                    </p>
                  </>
                )}
              </div>
            )}

            {!contentLoading && !permissions.canViewWorkbook && mode === 'workbook' && (
              <div className="accessInfoBox">
                Рабочая тетрадь доступна только обучающемуся
              </div>
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
                        <div className="dictCardMain">
                          <strong>{item.gesture_name}</strong>
                          <p>{item.gesture_description}</p>
                        </div>
                        <div className="dictCardMeta">
                          <span>{item.display_text}</span>
                          <small>{item.concept_name}</small>
                          <small>Слово: {item.word_name}</small>
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
                            <small>{mode === 'textbook' ? 'Режим учебника' : 'Самопроверка'}</small>
                          </div>

                          {mode === 'workbook' && permissions.canUseSelfCheck && (
                            <p className="exercisePrompt">
                              Запишите последовательность жестов для фразы
                            </p>
                          )}

                          <p className="phrase">{exercise.phrase}</p>

                          {mode === 'workbook' && permissions.canUseSelfCheck && (
                            <div className="workbookAnswerCard">
                              <span className="workbookAnswerLabel">Ваш ответ</span>
                              <textarea
                                className="workbookTextarea"
                                value={workbookAnswers[exercise.id] ?? ''}
                                onChange={(event) => updateWorkbookAnswer(exercise.id, event.target.value)}
                                placeholder="Введите последовательность жестов в удобной для вас записи"
                              />
                              <small className="workbookAnswerHint">
                                {showAnswer ? 'Образец последовательности' : 'Ответ скрыт для самопроверки'}
                              </small>
                            </div>
                          )}

                          {mode === 'workbook' && permissions.canUseSelfCheck && (
                            <button
                              className="answerButton"
                              onClick={() => toggleAnswer(exercise.id)}
                            >
                              {showAnswer ? 'Скрыть образец ответа' : 'Показать образец ответа'}
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
      )}
    </main>
  )
}

export default App
