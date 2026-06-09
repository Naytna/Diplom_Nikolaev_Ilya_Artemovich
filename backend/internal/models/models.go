package models

import "time"

type Exercise struct {
	ID           int64     `json:"id"`
	ThemeID      int64     `json:"theme_id"`
	LitExampleID int64     `json:"lit_example_id"`
	ExerciseType string    `json:"exercise_type"`
	TargetMode   string    `json:"target_mode"`
	Phrase       string    `json:"phrase"`
	Status       string    `json:"status"`
	Explanation  string    `json:"explanation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Segments     []ExerciseSegment `json:"segments" gorm:"foreignKey:ExerciseID"`
}

func (Exercise) TableName() string {
	return "learning.exercises"
}

type ExerciseSegment struct {
	ID               int64  `json:"id"`
	ExerciseID       int64  `json:"exercise_id"`
	SourceSegmentID  int64  `json:"source_segment_id"`
	TranslatedWordID int64  `json:"translated_word_id"`
	WordText         string `json:"word_text"`
	GestureName      string `json:"gesture_name"`
	VideoURL         string `json:"video_url"`
	AnswerVisible    bool   `json:"answer_visible"`
	PositionIndex    int    `json:"position_index"`
}

func (ExerciseSegment) TableName() string {
	return "learning.exercise_segments"
}

type GenerationRejection struct {
	ID              int64     `json:"id"`
	GenerationRunID int64    `json:"generation_run_id"`
	ThemeID         int64    `json:"theme_id"`
	LitExampleID    int64    `json:"lit_example_id"`
	ExampleText     string   `json:"example_text"`
	ReasonCode      string   `json:"reason_code"`
	ReasonText      string   `json:"reason_text"`
	CreatedAt       time.Time `json:"created_at"`
}

func (GenerationRejection) TableName() string {
	return "learning.generation_rejections"
}