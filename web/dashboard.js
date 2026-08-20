// cleanroom-monitor 看板逻辑：轮询 /api/v1/dashboard 渲染房间卡片。
(function () {
  "use strict";

  const roomsEl = document.getElementById("rooms");
  const roomCountEl = document.getElementById("room-count");
  const alertCountEl = document.getElementById("alert-count");
  const cacheTimeEl = document.getElementById("cache-time");
  const updatedAtEl = document.getElementById("updated-at");

  const STATUS_TEXT = {
    at_rest: "静态",
    normal: "正常",
    alert: "告警",
    alarm: "严重",
    restricted: "受限",
    release: "放行",
  };

  const PARAM_TEXT = {
    temp: "温度(℃)",
    humidity: "湿度(%)",
    pressure: "压差(Pa)",
    particle_05: "粒子≥0.5μm",
    particle_50: "粒子≥5.0μm",
  };

  async function refresh() {
    try {
      const resp = await fetch("/api/v1/dashboard");
      if (!resp.ok) throw new Error("HTTP " + resp.status);
      const data = await resp.json();
      render(data);
    } catch (e) {
      updatedAtEl.textContent = "加载失败: " + e.message;
    }
  }

  function render(data) {
    updatedAtEl.textContent = new Date(data.updated_at).toLocaleString("zh-CN");
    roomCountEl.textContent = data.rooms ? data.rooms.length : 0;
    alertCountEl.textContent = data.open_alerts || 0;
    cacheTimeEl.textContent = data.updated_at ? new Date(data.updated_at).toLocaleTimeString("zh-CN") : "-";

    const rooms = data.rooms || [];
    roomsEl.innerHTML = "";
    for (const room of rooms) {
      roomsEl.appendChild(renderRoom(room));
    }
  }

  function renderRoom(room) {
    const div = document.createElement("div");
    div.className = "room status-" + (room.status || "normal");

    const head = document.createElement("div");
    head.className = "room-head";

    const title = document.createElement("h3");
    title.textContent = room.room_code + " " + (STATUS_TEXT[room.status] || room.status);
    head.appendChild(title);

    const badge = document.createElement("span");
    badge.className = "badge" + (room.status === "alarm" ? " alarm" : room.status === "alert" ? " alert" : "");
    badge.textContent = "未决 " + room.open_alerts;
    head.appendChild(badge);

    div.appendChild(head);

    const readings = document.createElement("div");
    readings.className = "readings";
    const rt = room.realtime || [];
    if (rt.length === 0) {
      const empty = document.createElement("div");
      empty.className = "empty";
      empty.textContent = "暂无读数";
      readings.appendChild(empty);
    } else {
      for (const r of rt) {
        const row = document.createElement("div");
        row.className = "reading";
        const name = document.createElement("span");
        name.textContent = PARAM_TEXT[r.param_type] || r.param_type;
        const val = document.createElement("span");
        val.className = "val " + (r.within_range ? "ok" : "bad");
        val.textContent = r.value.toFixed(2);
        row.appendChild(name);
        row.appendChild(val);
        readings.appendChild(row);
      }
    }
    div.appendChild(readings);
    return div;
  }

  document.getElementById("refresh-btn").addEventListener("click", refresh);
  refresh();
  setInterval(refresh, 5000);
})();