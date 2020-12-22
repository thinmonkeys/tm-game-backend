package common

import db "../../store"

type ScoreGetter func(cif string) (db.DynamicScoreRecord, bool, error)
type ScorePutter func(record db.DynamicScoreRecord) error
type CategoryScoreGetAll func(cif string) ([]db.ScoreHistoryRecord, error)
type CategoryScoreGetter func(cif string, categoryCode string) (db.ScoreHistoryRecord, bool, error)
type CategoryScorePutter func(record db.ScoreHistoryRecord) error
type BadgeGetter func(cif string) ([]db.BadgeHistoryRecord, error)
type BadgePutter func(record db.BadgeHistoryRecord) error
type AllScoreGetter func() ([]int, error)