# Driver Certification Tests

Catalogue of certification tests used to validate that a driver built with this SDK
is production-ready. Source of truth: Confluence space **NMR → Drivers e
integraciones → Tests** (`https://netsocs.atlassian.net/wiki/spaces/NMR`).

Each test page follows a fixed format: **Objective / Test Steps / Expected Results /
Notes**. This document lists the **title, objective and applicability** of every test.

> All test titles, steps, expected results and audit-log strings are written in **English**.

---

## Applicability model

Which tests a driver must pass is determined by the **object types it creates**
(`ObjectMetadata.Type`). A driver only has to pass the tests for the object types it
actually exposes, plus all **Global** tests.

| Tag | Trigger — driver creates… | SDK object type |
|-----|---------------------------|-----------------|
| **Global** | any driver | (any) |
| **Video** | camera / NVR / DVR channels | `video_channel`, `video_engine` |
| **Alarm panel** | intrusion / security panels & zones | `alarm_panel` (+ zone `sensor`) |
| **Fire alarm** | fire detection panels & zones | `alarm_panel` / `sensor` (fire) |
| **Switch / Output** | relays, outputs, sirens | `switch` |
| **Notifier** | push / message / media notifications | `notifier` |
| **Door** | controllable doors / barriers | `door` |
| **Access control** | locks & credential readers | `lock`, `reader` |
| **Sensor** | measurements / detectors | `sensor` |
| **Log** | device historical logs | `log` |
| **Topology** | main device + peripherals (NVR↔cam, panel↔sensor) | parent/child objects |

---

## Existing tests

| # | Title | Applies to | Objective |
|---|-------|-----------|-----------|
| 001 | Duplicate Addresses | **Global** | Ensure that the driver prevents device duplication (two connections to the same device). |
| 002 | Communication Startup and Object Creation | **Global** | Verify the driver's behavior when communication starts and the objects needed for operation are created; the driver reports object count per domain and confirms communication started. |
| 003 | Misconfiguration | **Global** | Verify the driver's behavior with configuration values outside the defined ranges (addresses, extrafields), detecting errors during startup. |
| 004 | Invalid Credentials at Runtime | **Global** (credential-based devices) | Verify the driver detects invalid credentials (password / username / encryption key) at runtime. |
| 005 | Lost Connection | **Global** | Verify the driver's behavior when the connection drops between the Netsocs server and the device. |
| 006 | Lost Communication | **Topology** | Verify the driver's behavior when communication is lost between a main device and its peripherals (camera↔NVR, sensor↔receiver). |
| 007 | Credential Change from Netsocs | **Global** (credential-based devices) | Verify the driver's behavior when device credentials are changed from Netsocs and a reconnection is forced. |
| 008 | Device Log Object(s) Integration | **Log** | Verify the driver's behavior when viewing the historical log of `log` type objects. |
| 009 | Disconnection During Live Video Streaming | **Video** | Verify the viewer's behavior when the live stream connection is interrupted. |
| 010 | Disconnection During Recorded Video Streaming | **Video** | Verify the viewer's behavior when the recorded stream connection is interrupted. |
| 013 | Alarm Event Processing Test for Security Systems | **Fire alarm** / **Alarm panel** | Verify driver operation for fire alarm devices or similar. ⚠️ *Title says "Security Systems" but body is fire-alarm — near-duplicate of 014.* |
| 014 | Fire Alarm Device Behavior Test | **Fire alarm** | Verify driver operation for fire alarm devices or similar. |
| 015 | Tamper | **Alarm panel** / **Access control** | Verify behavior when processing tampering events in access/alarm systems. |
| 021 | Battery Failure | **Alarm panel** | Verify event handling on battery / backup-power failure. |
| 053 | Live Video Streaming | **Video** | Verify live video streaming from cameras/DVR/encoders. |
| 055 | Multiple Streams from One Source | **Video** | Verify simultaneous live streaming of multiple videos from a single source. |
| 057 | Playback from Specific Time | **Video** | Verify recorded-video playback from a specific time. |
| 058 | Playback with No Available Recording | **Video** | Verify recorded-video behavior when no recording exists at the requested time. |
| 059 | Playback from Multiple Sources | **Video** | Verify simultaneous recorded-video playback from multiple sources. |
| 060 | Video Recording Controls | **Video** | Verify recorded-stream controls (pause/resume, fast-forward, etc.). |
| 061 | Stream Control from Multiple Locations | **Video** | Verify recorded-playback control from multiple locations simultaneously. |
| 062 | Track Change During Playback | **Video** | ⚠️ *Body is an exact duplicate of 061; does not test track change. Needs rewrite.* |
| 064 | Recorded Video Viewing Test with Time Zone Support | **Video** | Verify recorded-video playback with correct time-zone handling between user and device. |
| 068 | Area or Partition Arming/Disarming | **Alarm panel** | Verify arming/disarming handling from device and Netsocs UI. |
| 069 | Area Arming Failure | **Alarm panel** | Verify behavior when arming fails (detector active at arm time). |
| 070 | Zone Bypass and Unbypass | **Alarm panel** / **Fire alarm** | Verify detector bypass/unbypass compatibility. |
| 071 | Zone Test Mode Activation | **Alarm panel** / **Fire alarm** | Verify test-mode compatibility (`__test_mode` property + audit logs) from UI and panel. |
| 082 | Video Functions - Recording Download | **Video** | Verify download/storage of selected recordings to a local client drive. |
| 083 | PTZ | **Video** (PTZ-capable) | Verify PTZ capability (position, zoom, focus, iris). |
| 084 | Output Activation and Deactivation | **Switch / Output** | Verify activation/deactivation of outputs (`switch` objects). |
| 085 | `video_channel` Object Properties | **Video** | Verify `video_channel` objects expose properties identifying channel capabilities. |
| 086 | Event Reception | **Global** (event-emitting objects) | Verify the driver sends events with the properties defined in the manual. |
| 087 | Analytics Reception | **Video** (analytics-capable) | Verify reception of analytics (typed objects / analytics events, bounding boxes). ⚠️ *Objective on the page is wrongly copied from 084.* |
| 088 | Intrusion Alarm Reception | **Alarm panel** (intrusion) | Verify processing of intrusion-alarm events. |
| 089 | Ping Tests | **Global** | Verify connectivity check to a remote device before it is added. |
| 090 | Password Change with Driver Offline | **Global** (credential-based devices) | Verify the driver detects a remote-device password change that happened while offline. |
| 091 | Snapshot Capture for Video Objects | **Video** | Verify the snapshot method returns a still image of the current channel on demand. |
| 092 | Video Push (H.264 and H.265) | **Video** | Verify the driver pushes live video in H.264 and H.265 (HEVC) and reports the codec. |
| 093 | Object ID Uniqueness Across Devices | **Global** | Ensure object IDs never collide across different devices of the same driver. |
| 094 | Driver Lifecycle: Install, Restart, Update, Auto-start | **Global** | Verify install, manual restart, in-place update, and auto-start after a site reboot. |
| 095 | Notification Delivery | **Notifier** | Verify `create` delivers a notification to the target and reflects idle/busy/error states. |
| 096 | Notification Rich Media and Deduplication | **Notifier** | Verify rich-media payloads (image/audio/video URLs) and dedup by `notification_id`. |
| 097 | Door Open and Close Command | **Door** | Verify `door.action.open`/`close` execute and set `door.state.open`/`close` + audit log. |
| 098 | Door Physical State Reflection and Forced/Held-Open | **Door** | Verify object reflects physical open/close/lock/transient states and forced/held-open events. |
| 099 | Reader Credential Read and Supported Types | **Reader** | Verify `read` per supported type, rejects unsupported type, publishes `supported_credential_types`. |
| 100 | Reader Live Access Event Reception | **Reader** | Verify access events (person, credential, granted/denied) on credential presentation. |
| 101 | Reader Access Database Full Sync | **Reader** | Verify `sync_access_database` full = complete replacement, atomic, idempotent. |
| 102 | Reader Access Database Incremental Sync | **Reader** | Verify incremental upsert + `deleted_ids`, untouched persons preserved, empty = no-op. |
| 103 | Reader Automatic Sync Triggers and 2-Minute Timeout | **Reader** | Verify driver_connect/heartbeat/event triggers and confirm within 2-min timeout. |
| 104 | Reader Person Restrictions (Disabled/Validity/Holidays/Bands) | **Reader** | Verify `enabled`, `valid_from/until`, `holidays`, `bands` enforced or documented. |
| 105 | Reader Anti-Passback (APB) | **Reader** (APB hardware) | Verify `apb_area` none/soft/hard, direction, `apb_exempt`; ignore if unsupported. |
| 106 | Reader Credential Management Actions | **Reader** | Verify `get_people`/`set_people`/`delete_person`/`store_qrs`/`delete_qrs`, multi-credential, ignore unsupported. |
| 107 | Reader Concurrency and Idempotency | **Reader** | Verify mutex-protected concurrent actions and idempotent upsert/delete. |

> Numbering has gaps (011, 012, 016–020, 022–052, 054, 056, 063, 065–067, 072–081).
> Confirm whether these are reserved or deleted before reusing.

---

## Required tests by driver type

Every driver must pass the **Global** set. Add the rows for each object type the driver creates.

### Global (all drivers)
001, 002, 003, 004, 005, 007, 086, 089, 090, 093, 094
*(006 if the device has peripherals; 008 if it creates `log` objects.)*

### Video driver (`video_channel` / `video_engine`)
009, 010, 053, 055, 057, 058, 059, 060, 061, 062, 064, 082, 085, 091, 092
*(083 if PTZ-capable; 087 if analytics-capable.)*

### Alarm panel / intrusion driver (`alarm_panel`)
015, 021, 068, 069, 070, 071, 088
*(006 panel↔sensor topology.)*

### Fire alarm driver (`alarm_panel` / fire `sensor`)
013, 014, 070, 071

### Switch / output driver (`switch`)
084

### Notifier driver (`notifier`)
095, 096

### Door driver (`door`)
097, 098

### Access control / Reader driver (`reader`)
099, 100, 101, 102, 103, 104, 106, 107 *(105 if APB hardware; 015 tamper if it also exposes `lock`/`reader` tamper)*

### Access control driver (`lock`)
015 *(tamper; plus Global.)*

### Sensor driver (`sensor`)
*(Global only; sensor readings validated via 086 Event Reception and per-object manual.)*

### Log driver (`log`)
008

---

## Known defects (to fix in Confluence)

1. **087 Analytics Reception** — Objective text wrongly copied from 084 ("activating and deactivating outputs"); body is correct.
2. **062 Track Change During Playback** — body is an exact duplicate of 061; does not test track change.
3. **013 / 014** — near-duplicates; 013 is titled "Security Systems" but its content is fire-alarm.
