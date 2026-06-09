package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type GenerationService struct {
	db *gorm.DB
}

type GenerationResult struct {
	RunID              int64  `json:"run_id"`
	ThemeID            int64  `json:"theme_id"`
	FoundExamples      int    `json:"found_examples"`
	GeneratedExercises int    `json:"generated_exercises"`
	RejectedExamples   int    `json:"rejected_examples"`
	DurationMS         int64  `json:"duration_ms"`
	Status             string `json:"status"`
}

type candidateExample struct {
	ID   int64
	Text string
}

type sourceSegment struct {
	ID               int64
	LitExampleID     int64
	TranslatedWordID int64
	WordText         string
	GestureName      string
	VideoURL         string
	PositionIndex    int
}

type generationRejection struct {
	ThemeID      int64
	LitExampleID int64
	ExampleText  string
	ReasonCode   string
	ReasonText   string
}

func NewGenerationService(db *gorm.DB) *GenerationService {
	return &GenerationService{db: db}
}

func (s *GenerationService) Generate(themeID int64, userID int64) (GenerationResult, error) {
	startedAt := time.Now()
	result := GenerationResult{ThemeID: themeID, Status: "failed"}
	rejections := make([]generationRejection, 0)

	themeWords, err := s.loadThemeTranslatedWords(themeID)
	if err != nil {
		_, _ = s.saveRun(themeID, userID, result, err.Error(), startedAt)
		return result, err
	}
	if len(themeWords) < 3 {
		err = errors.New("для генерации нужно минимум 3 переводных слова в теме")
		_, _ = s.saveRun(themeID, userID, result, err.Error(), startedAt)
		return result, err
	}

	candidates, err := s.findCandidateExamples(themeWords)
	if err != nil {
		_, _ = s.saveRun(themeID, userID, result, err.Error(), startedAt)
		return result, err
	}
	result.FoundExamples = len(candidates)

	themeSet := make(map[int64]bool)
	for _, id := range themeWords {
		themeSet[id] = true
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, example := range candidates {
			segments, err := s.loadExampleSegments(tx, example.ID)
			if err != nil {
				return err
			}

			allowed, reasonCode, reasonText := s.validateExample(segments, themeSet)
			if !allowed {
				result.RejectedExamples++
				rejections = append(rejections, generationRejection{
					ThemeID:      themeID,
					LitExampleID: example.ID,
					ExampleText:  example.Text,
					ReasonCode:   reasonCode,
					ReasonText:   reasonText,
				})
				continue
			}

			used, err := s.exampleAlreadyUsed(tx, themeID, example.ID)
			if err != nil {
				return err
			}
			if used {
				result.RejectedExamples++
				rejections = append(rejections, generationRejection{
					ThemeID:      themeID,
					LitExampleID: example.ID,
					ExampleText:  example.Text,
					ReasonCode:   "already_used",
					ReasonText:   "Пример уже использовался для генерации упражнений в выбранной теме",
				})
				continue
			}

			created, err := s.createExercisePair(tx, themeID, example, segments)
			if err != nil {
				return err
			}
			result.GeneratedExercises += created
		}
		return nil
	})
	if err != nil {
		_, _ = s.saveRun(themeID, userID, result, err.Error(), startedAt)
		return result, err
	}

	result.DurationMS = time.Since(startedAt).Milliseconds()
	result.Status = "completed"

	runID, err := s.saveRun(themeID, userID, result, "", startedAt)
	if err != nil {
		return result, err
	}
	result.RunID = runID

	if len(rejections) > 0 {
		if err := s.saveRejections(runID, rejections); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (s *GenerationService) loadThemeTranslatedWords(themeID int64) ([]int64, error) {
	var ids []int64
	err := s.db.Table("learning.theme_translated_words").
		Where("theme_id = ?", themeID).
		Pluck("translated_word_id", &ids).Error
	return ids, err
}

func (s *GenerationService) findCandidateExamples(ids []int64) ([]candidateExample, error) {
	var examples []candidateExample
	err := s.db.Raw(`
		select distinct le.id, le.text
		from linguistic.lit_examples le
		join linguistic.lit_example_segments les on les.lit_example_id = le.id
		where le.status in ('verified', 'published')
		and les.translated_word_id in ?
		order by le.id
	`, ids).Scan(&examples).Error
	return examples, err
}

func (s *GenerationService) loadExampleSegments(tx *gorm.DB, exampleID int64) ([]sourceSegment, error) {
	var segments []sourceSegment
	err := tx.Raw(`
		select
			les.id,
			les.lit_example_id,
			les.translated_word_id,
			les.text as word_text,
			g.name as gesture_name,
			coalesce(g.video_url, '') as video_url,
			les.position_index
		from linguistic.lit_example_segments les
		join linguistic.translated_words tw on tw.id = les.translated_word_id
		join linguistic.gestures g on g.id = tw.gesture_id
		where les.lit_example_id = ?
		order by les.position_index
	`, exampleID).Scan(&segments).Error
	return segments, err
}

func (s *GenerationService) validateExample(segments []sourceSegment, themeSet map[int64]bool) (bool, string, string) {
	if len(segments) < 2 {
		return false, "not_enough_segments", "В примере меньше двух жестовых сегментов, поэтому из него нельзя сформировать полноценное упражнение"
	}

	outsideWords := make([]string, 0)
	for _, segment := range segments {
		if !themeSet[segment.TranslatedWordID] {
			label := segment.WordText
			if label == "" {
				label = segment.GestureName
			}
			outsideWords = append(outsideWords, label)
		}
	}

	if len(outsideWords) > 0 {
		return false,
			"outside_theme_vocabulary",
			fmt.Sprintf("Пример содержит слова вне словаря выбранной темы: %s", strings.Join(outsideWords, ", "))
	}

	return true, "", ""
}

func (s *GenerationService) exampleAlreadyUsed(tx *gorm.DB, themeID int64, exampleID int64) (bool, error) {
	var count int64
	err := tx.Table("learning.exercises").
		Where("theme_id = ? and lit_example_id = ?", themeID, exampleID).
		Count(&count).Error
	return count > 0, err
}

func (s *GenerationService) createExercisePair(tx *gorm.DB, themeID int64, example candidateExample, segments []sourceSegment) (int, error) {
	created := 0
	modes := []struct {
		TargetMode    string
		AnswerVisible bool
	}{
		{TargetMode: "textbook", AnswerVisible: true},
		{TargetMode: "workbook", AnswerVisible: false},
	}

	for _, mode := range modes {
		var exerciseID int64
		explanation := s.buildExplanation(segments)
		err := tx.Raw(`
			insert into learning.exercises
			(theme_id, lit_example_id, exercise_type, target_mode, phrase, status, explanation)
			values (?, ?, 'translation_sequence', ?, ?, 'draft', ?)
			returning id
		`, themeID, example.ID, mode.TargetMode, example.Text, explanation).Scan(&exerciseID).Error
		if err != nil {
			return created, err
		}

		for _, segment := range segments {
			err = tx.Exec(`
				insert into learning.exercise_segments
				(exercise_id, source_segment_id, translated_word_id, word_text, gesture_name, video_url, answer_visible, position_index)
				values (?, ?, ?, ?, ?, ?, ?, ?)
			`, exerciseID, segment.ID, segment.TranslatedWordID, segment.WordText, segment.GestureName, segment.VideoURL, mode.AnswerVisible, segment.PositionIndex).Error
			if err != nil {
				return created, err
			}
		}

		created++
	}

	return created, nil
}

func (s *GenerationService) buildExplanation(segments []sourceSegment) string {
	value := "Последовательность жестов: "
	for index, segment := range segments {
		if index > 0 {
			value += " "
		}
		value += fmt.Sprintf("[%s]", segment.GestureName)
	}
	return value
}

func (s *GenerationService) saveRun(themeID int64, userID int64, result GenerationResult, message string, startedAt time.Time) (int64, error) {
	duration := time.Since(startedAt).Milliseconds()
	var runID int64

	err := s.db.Raw(`
		insert into learning.generation_runs
		(theme_id, started_by, status, found_examples, generated_exercises, rejected_examples, duration_ms, error_message)
		values (?, ?, ?, ?, ?, ?, ?, ?)
		returning id
	`, themeID, userID, result.Status, result.FoundExamples, result.GeneratedExercises, result.RejectedExamples, duration, message).Scan(&runID).Error

	return runID, err
}

func (s *GenerationService) saveRejections(runID int64, rejections []generationRejection) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, rejection := range rejections {
			err := tx.Exec(`
				insert into learning.generation_rejections
				(generation_run_id, theme_id, lit_example_id, example_text, reason_code, reason_text)
				values (?, ?, ?, ?, ?, ?)
			`, runID, rejection.ThemeID, rejection.LitExampleID, rejection.ExampleText, rejection.ReasonCode, rejection.ReasonText).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}