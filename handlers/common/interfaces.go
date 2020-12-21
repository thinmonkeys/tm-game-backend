package common

import db "../../store"

type ScoreGetter func(cif string) (db.DynamicScoreRecord, bool, error)
type ScorePutter func(record db.DynamicScoreRecord) error