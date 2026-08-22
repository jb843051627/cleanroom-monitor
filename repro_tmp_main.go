//go:build ignore

// Reproduce bug-018: deleted room row + surviving batch row -> nil room panic.
// Run: go run -tags=ignore repro_tmp_main.go  (build tag keeps it out of normal build/test)
// This file lives at module root so internal/ packages are importable.
// It is NOT a _test.go file and is not compiled by default build/test.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cleanroom-monitor/internal/config"
	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/service"
	"cleanroom-monitor/internal/store"
)

func must(err error) {
	if err != nil {
		fmt.Println("FAIL setup:", err)
		os.Exit(1)
	}
}

func main() {
	dbPath := "repro_tmp.db"
	_ = os.Remove(dbPath)
	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	defer db.Close()

	svcs := service.NewServices(db, config.New())
	ctx := context.Background()

	crStore := store.NewCleanroomStore(db)
	rmStore := store.NewRoomStore(db)

	cr := &model.Cleanroom{Name: "CR1", Code: "C1", Grade: "B", AreaSqm: 10, Status: model.StateAtRest, CreatedAt: time.Now()}
	must(crStore.Create(ctx, cr))
	r := &model.Room{CleanroomID: cr.ID, Name: "R1", Code: "R1", Kind: "production", AreaSqm: 10, TargetPressure: 10, PressureTolerance: 1, TargetTemp: 22, TargetHumidity: 50, Status: model.StateAtRest, CreatedAt: time.Now()}
	must(rmStore.Create(ctx, r))

	b, err := svcs.Batches.Start(ctx, &model.BatchInput{RoomID: r.ID, Name: "B1", Product: "P", Phase: "fill"})
	if err != nil {
		fmt.Println("start:", err)
		os.Exit(1)
	}
	fmt.Println("batch started: id=", b.ID, "status=", b.Status)

	// Simulate "clear historical cleanroom": delete the room row.
	// clean_batches has no FK to rooms -> row survives, batch stays in history.
	_, err = db.ExecContext(ctx, "DELETE FROM rooms WHERE id=?", r.ID)
	must(err)

	rows, _ := db.QueryContext(ctx, "SELECT id,status,room_id FROM clean_batches ORDER BY id")
	for rows.Next() {
		var id, rid int64
		var st string
		rows.Scan(&id, &st, &rid)
		fmt.Printf("batch row: id=%d status=%s room_id=%d (room row now gone)\n", id, st, rid)
	}
	rows.Close()

	// Replay: Complete the (still in_progress) batch -> expect panic in transition.
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("PANIC REPRODUCED:", rec)
			os.Exit(0)
		}
	}()
	bb, err := svcs.Batches.Complete(ctx, b.ID)
	fmt.Println("complete returned:", bb, "err:", err)
}
