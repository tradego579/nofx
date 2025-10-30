package db

import (
	"context"
	"nofx/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName = "nofx"
	tradersColl  = "traders"
)

type TraderDoc struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	TraderID           string             `bson:"trader_id" json:"trader_id"`
	Name               string             `bson:"name" json:"name"`
	AIModel            string             `bson:"ai_model" json:"ai_model"`
	Exchange           string             `bson:"exchange" json:"exchange"`
	BinanceAPIKey      string             `bson:"binance_api_key,omitempty" json:"binance_api_key,omitempty"`
	BinanceSecretKey   string             `bson:"binance_secret_key,omitempty" json:"binance_secret_key,omitempty"`
	BinanceTestnet     bool               `bson:"binance_testnet,omitempty" json:"binance_testnet,omitempty"`
	HyperliquidPrivate string             `bson:"hyperliquid_private_key,omitempty" json:"hyperliquid_private_key,omitempty"`
	HyperliquidTestnet bool               `bson:"hyperliquid_testnet,omitempty" json:"hyperliquid_testnet,omitempty"`
	AsterUser          string             `bson:"aster_user,omitempty" json:"aster_user,omitempty"`
	AsterSigner        string             `bson:"aster_signer,omitempty" json:"aster_signer,omitempty"`
	AsterPrivateKey    string             `bson:"aster_private_key,omitempty" json:"aster_private_key,omitempty"`
	QwenKey            string             `bson:"qwen_key,omitempty" json:"qwen_key,omitempty"`
	DeepSeekKey        string             `bson:"deepseek_key,omitempty" json:"deepseek_key,omitempty"`
	CustomAPIURL       string             `bson:"custom_api_url,omitempty" json:"custom_api_url,omitempty"`
	CustomAPIKey       string             `bson:"custom_api_key,omitempty" json:"custom_api_key,omitempty"`
	CustomModelName    string             `bson:"custom_model_name,omitempty" json:"custom_model_name,omitempty"`
	InitialBalance     float64            `bson:"initial_balance" json:"initial_balance"`
	ScanIntervalMin    int                `bson:"scan_interval_minutes" json:"scan_interval_minutes"`
	Enabled            bool               `bson:"enabled" json:"enabled"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

func ToConfig(td TraderDoc) config.TraderConfig {
	return config.TraderConfig{
		ID:                    td.TraderID,
		Name:                  td.Name,
		AIModel:               td.AIModel,
		Exchange:              td.Exchange,
		BinanceAPIKey:         td.BinanceAPIKey,
		BinanceSecretKey:      td.BinanceSecretKey,
		BinanceTestnet:        td.BinanceTestnet,
		HyperliquidPrivateKey: td.HyperliquidPrivate,
		HyperliquidTestnet:    td.HyperliquidTestnet,
		AsterUser:             td.AsterUser,
		AsterSigner:           td.AsterSigner,
		AsterPrivateKey:       td.AsterPrivateKey,
		QwenKey:               td.QwenKey,
		DeepSeekKey:           td.DeepSeekKey,
		CustomAPIURL:          td.CustomAPIURL,
		CustomAPIKey:          td.CustomAPIKey,
		CustomModelName:       td.CustomModelName,
		InitialBalance:        td.InitialBalance,
		ScanIntervalMinutes:   td.ScanIntervalMin,
		Enabled:               td.Enabled,
	}
}

func FromConfig(tc config.TraderConfig) TraderDoc {
	now := time.Now()
	return TraderDoc{
		TraderID:           tc.ID,
		Name:               tc.Name,
		AIModel:            tc.AIModel,
		Exchange:           tc.Exchange,
		BinanceAPIKey:      tc.BinanceAPIKey,
		BinanceSecretKey:   tc.BinanceSecretKey,
		BinanceTestnet:     tc.BinanceTestnet,
		HyperliquidPrivate: tc.HyperliquidPrivateKey,
		HyperliquidTestnet: tc.HyperliquidTestnet,
		AsterUser:          tc.AsterUser,
		AsterSigner:        tc.AsterSigner,
		AsterPrivateKey:    tc.AsterPrivateKey,
		QwenKey:            tc.QwenKey,
		DeepSeekKey:        tc.DeepSeekKey,
		CustomAPIURL:       tc.CustomAPIURL,
		CustomAPIKey:       tc.CustomAPIKey,
		CustomModelName:    tc.CustomModelName,
		InitialBalance:     tc.InitialBalance,
		ScanIntervalMin:    tc.ScanIntervalMinutes,
		Enabled:            tc.Enabled,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func ListTraders(ctx context.Context) ([]TraderDoc, error) {
	cli, err := Connect(ctx)
	if err != nil {
		return nil, err
	}
	col := cli.Database(databaseName).Collection(tradersColl)
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []TraderDoc
	for cur.Next(ctx) {
		var d TraderDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, cur.Err()
}

func UpsertTrader(ctx context.Context, td TraderDoc) error {
	cli, err := Connect(ctx)
	if err != nil {
		return err
	}
	col := cli.Database(databaseName).Collection(tradersColl)
	// 构建$set文档，排除created_at
	td.UpdatedAt = time.Now()
	setDoc := bson.M{
		"trader_id":               td.TraderID,
		"name":                    td.Name,
		"ai_model":                td.AIModel,
		"exchange":                td.Exchange,
		"binance_api_key":         td.BinanceAPIKey,
		"binance_secret_key":      td.BinanceSecretKey,
		"binance_testnet":         td.BinanceTestnet,
		"hyperliquid_private_key": td.HyperliquidPrivate,
		"hyperliquid_testnet":     td.HyperliquidTestnet,
		"aster_user":              td.AsterUser,
		"aster_signer":            td.AsterSigner,
		"aster_private_key":       td.AsterPrivateKey,
		"qwen_key":                td.QwenKey,
		"deepseek_key":            td.DeepSeekKey,
		"custom_api_url":          td.CustomAPIURL,
		"custom_api_key":          td.CustomAPIKey,
		"custom_model_name":       td.CustomModelName,
		"initial_balance":         td.InitialBalance,
		"scan_interval_minutes":   td.ScanIntervalMin,
		"enabled":                 td.Enabled,
		"updated_at":              td.UpdatedAt,
	}
	_, err = col.UpdateOne(ctx,
		bson.M{"trader_id": td.TraderID},
		bson.M{"$set": setDoc, "$setOnInsert": bson.M{"created_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}

func DeleteTrader(ctx context.Context, traderID string) error {
	cli, err := Connect(ctx)
	if err != nil {
		return err
	}
	col := cli.Database(databaseName).Collection(tradersColl)
	_, err = col.DeleteOne(ctx, bson.M{"trader_id": traderID})
	return err
}

// no extra helpers
