# Family Feud Backend: TypeScript to Go Migration Plan

This plan describes how to migrate the current Node.js/TypeScript backend to Go while preserving existing clients, MongoDB data, Redis live-game state, and Socket.IO gameplay behavior. It is based on the current `src/` tree on the `multiplayer` branch, including the refresh-token cookie migration.

## Current Implementation Baseline

### Runtime and startup

- Node.js, TypeScript, Express 5, native HTTP server, Mongoose 9, and dotenv.
- `src/main.ts` serves `public/`, enables CORS with credentials, parses JSON and URL-encoded bodies, logs with Morgan, connects to MongoDB, initializes Socket.IO, and exposes `/`, `/live`, and `/health`.
- `/health` is a MongoDB readiness check only. Redis is required by Socket.IO initialization and live state, but is not represented in the health response.
- MongoDB connection settings use `autoIndex`, a 5-second server-selection timeout, and a 45-second socket timeout.
- Socket.IO initialization is asynchronous and connects two Redis adapter clients. Startup sequencing and shutdown ownership must be made explicit before production cutover.

### Current source layout

- `src/config/`: MongoDB and Socket.IO setup.
- `src/middleware/`: HTTP authentication/authorization, request-data extraction, and socket authentication.
- `src/modules/auth/`: registration, login, cookie-based refresh rotation, logout, and password changes.
- `src/modules/user/`, `game/`, `game_session/`, `survey_details/`, `questions/`, and `ai/`: HTTP controllers, services, DTOs, routes, and schemas.
- `src/modules/game_live/`: live-game and Fast Money state, event persistence models, and live-game services.
- `src/modules/game_machine/`: Socket.IO handlers, room helpers, state-machine rules, latency compensation, and game DTOs.
- `src/utils/`: response/error handling, enums/constants, pagination, providers, Redis state, persistence queues, transactions, and the base schema plugin.

### HTTP route contract

All routes below are mounted below `/api`.

| Method | Path | Auth | Responsibility |
| --- | --- | --- | --- |
| POST | `/auth/register` | public | Register a user |
| POST | `/auth/login` | public | Login, return an access token, and set the HTTP-only refresh cookie |
| POST | `/auth/refresh-token` | `refresh_token` cookie | Rotate the refresh cookie and return a new access token |
| POST | `/auth/logout` | access token | Logout and clear refresh cookie |
| POST | `/auth/change-password` | access token | Change password |
| GET/PATCH/DELETE | `/users/:id` | access token | Read, update, or soft-delete a user |
| PATCH | `/users/:id/restore` | access token | Restore a user |
| POST | `/games` | access token + host role | Create a game |
| GET | `/games` | access token | List games |
| GET | `/games/:id` | access token | Fetch a game |
| PATCH | `/games/:id` | access token + owner | Update a game |
| DELETE | `/games/:id` | access token + owner | Soft-delete a game |
| PATCH | `/games/:id/restore` | access token + owner | Restore a game |
| PATCH | `/games/:id/update-status` | access token + owner | Change game status |
| POST | `/survey` | access token | Create a survey |
| GET | `/survey/game/:gameId` | access token | Fetch a game's survey |
| POST | `/survey/response/submit` | public | Submit a survey response |
| PATCH | `/survey/response` | access token | Update a survey response |
| DELETE | `/survey/response` | access token | Delete a survey response |
| GET | `/survey/:code/code` | public | Fetch a survey by public code |
| PATCH | `/survey/:id/status` | access token | Change survey status |
| PATCH | `/survey/:id` | access token | Update a survey |
| POST | `/game-session` | access token | Create a session |
| GET | `/game-session` | access token | List sessions |
| GET | `/game-session/:code/code` | public | Fetch a session by join code |
| POST | `/game-session/:code/join` | public | Join a session as a player |
| GET | `/game-session/:code/rejoin` | public | Rejoin using device identity |
| PATCH | `/game-session/:id/team/name` | access token + host | Rename a team |
| PATCH | `/game-session/:id/add/team` | access token + host | Add a team |
| PATCH | `/game-session/:id/remove/team` | access token + host | Remove a team |
| PATCH | `/game-session/:id/add/player` | access token + host | Add a player |
| PATCH | `/game-session/:id/update/team/player` | access token + host | Update a player |
| PATCH | `/game-session/:id/remove/team/player` | access token + host | Remove a player |
| PATCH | `/game-session/:id/unban/device` | access token + host | Unban a device |
| PATCH | `/game-session/:id/update/player/role` | access token + host | Change captain/member role |
| PATCH | `/game-session/:id/update-status` | access token + host | Change session status |
| PATCH | `/game-session/:id/settings` | access token + host | Update session settings |
| PATCH | `/questions/:questionId/answers` | access token | Update answer slots |
| POST | `/ai/generate-questions` | access token + host role | Generate questions |
| GET | `/ai/analysis/survey/response/:id` | access token + host role | Analyze survey responses |

The `/survey` and `/game-session` routers are mounted without a global auth middleware; their individual routes define access. This distinction is part of the compatibility contract.

"access token + owner"/"+ host" means the route requires both a valid access token and that the authenticated user is the resource's owner: `Game.createdBy` for `/games/:id/*` routes (checked by `assertGameOwner`), and `GameSession.hostId` for `/game-session/:id/*` routes (checked by `assertSessionHost`). Both middlewares 404 on a missing resource and 403 (`FORBIDDEN`) when the authenticated user is not the owner/host, even though the token itself is valid. Go must reproduce this two-step check (identity, then ownership) rather than collapsing it into a single role check.

Success responses use `{ success: true, message, data }`. Login and refresh `data` contain only `accessToken`; the refresh token is never returned in JSON. Paginated responses use `{ success: true, message, metadata, data }`. Errors use `{ success: false, message, code?, error? }`; development HTTP errors may include `stack`. Preserve status codes, null values, ObjectId strings, pagination metadata, and error codes.

### Authentication and authorization

- HTTP access tokens are verified from `Authorization: Bearer <token>` and expose `id` and `role` claims.
- Refresh tokens use the `refresh_token` HTTP-only cookie. Cookie path defaults to `/api/auth`; production uses `secure` and `SameSite=None`, while development uses `SameSite=Lax`. Optional `COOKIE_DOMAIN` is supported.
- Credentialed production CORS requires explicit `ALLOWED_ORIGINS`; wildcard origins are rejected during configuration validation. Client requests must include credentials so browsers and supported native clients send the cookie.
- Passwords use Argon2. Existing hashes must validate in Go.
- Roles are `host`, `guest`, and `player`; host-only operations are enforced by middleware.
- Team sockets use a separate player token signed with the access secret. It contains `typ: "player"`, `sessionId`, `teamId`, `playerId`, and `deviceId`, and expires after 12 hours.
- Socket handshake accepts `auth` first and query parameters as fallback: `sessionId`, `role`, `teamId`, `gameliveId`, team `playerToken`, and moderator `accessToken`. The server verifies player-token session/team binding and derives team/player/device identity from the token.
- A `role: "moderator"` connection must include a valid `accessToken` (the same JWT issued at login). The server verifies the JWT, then confirms the decoded user ID equals `GameSession.hostId` for the target `sessionId` (`GameSessionService.isHost`), rejecting the handshake with an `Error` (missing/invalid token, or "not the host") otherwise. `role: "audience"` remains unauthenticated - only `team` and `moderator` connections are identity-checked. Go must reproduce this per-role handshake branching, not a single shared check.

### MongoDB collections and data rules

Mongoose models and collection contracts are:

- `User`
- `Game`
- `GameSession`
- `Survey`
- `SurveyRespondent`
- `Questions`
- `GameLive`
- `GameLiveEvent`

There are eight persisted models/collections in the current backend, not seven.

Every schema using the base plugin has `createdAt`, `updatedAt`, `createdBy`, `updatedBy`, `isDeleted`, and `deletedAt`. Normal `find` queries exclude deleted documents unless `withDeleted` is explicitly set. Delete and restore are soft operations. Preserve schema defaults, enum validation, nested ObjectIds, timestamps, indexes, unique constraints, population behavior, and transaction behavior.

`GameLive` is unique by `sessionId` and contains versioned live state: rounds, answer slots, buzzer presses and nominations, face-off attempts and decisions, scoreboard, disqualified teams, winner/tie/result flags, prizes, and optional Fast Money state. Fast Money contains two ordered players, five-question responses, duplicate/pass/reveal flags, per-player clocks, pause accounting, target score, and phase transitions.

`GameLiveEvent` is an append-only audit log with event type, live-game/session/round/team/player identifiers, and a nullable payload containing answer, correctness, points, strike, decision, state, question, and duplicate fields. Event enums and field names are externally relevant for history consumers.

### Redis contract

The state client uses `REDIS_URL` and the following keys:

| Key | Purpose | Expiry/behavior |
| --- | --- | --- |
| `game:<sessionId>` | Authoritative live state JSON | 24 hours; initial Mongo seed uses `NX` |
| `lock:game:<sessionId>` | Per-session distributed mutation lock | `SET NX PX`, default 3 seconds; release is token-checked Lua compare/delete |
| `buzz:<sessionId>:<roundId>` | Cross-instance buzzer winner lock | `SET NX PX`, default 3 seconds |
| `roster:<sessionId>` | Roster cache | Preserve existing callers and TTL policy |
| `gamedef:<gameId>` | Game definition cache | Preserve existing callers and TTL policy |
| `surveydef:<surveyId>` | Survey definition cache | Preserve existing callers and TTL policy |
| `surveycode:<UPPERCASE_CODE>` | Public survey-code cache | Preserve uppercase normalization |
| `presence:<sessionId>:teams` | Team presence hash | 24-hour safety TTL |
| `presence:<sessionId>:moderator/audience` | Moderator/audience counters | 24-hour safety TTL |

Live mutations must remain: acquire lock, load Redis state, validate and mutate, increment `version`, save state, publish/broadcast, enqueue persistence, release lock. MongoDB work must remain outside the critical Redis lock section.

### Realtime contract

Socket.IO uses websocket and polling transports, a 60-second ping timeout, a 25-second ping interval, and the Redis adapter for multi-instance broadcasts. Rooms are:

```text
mod_<sessionId>
pub_<sessionId>
team_<sessionId>_<teamId>
teams_<sessionId>
```

The protocol includes lobby/state synchronization, presence, moderator round controls, face-off/buzzer controls, answer judging, strikes/steals, team disqualification, player removal, game abort/completion, and Fast Money. The complete event names are defined in `src/utils/common/constant.ts` and handlers in `src/modules/game_machine/register.game.handlers.ts` and `register.fast.money.handlers.ts`; migration tests must snapshot that list rather than manually maintaining a shortened list.

Important behavior to preserve:

- Moderator, team, and audience authorization is checked per event. Moderator-only events (`assertModerator`) trust `socket.data.role`, which is only set to `"moderator"` after the handshake's host-identity check above passes - per-event handlers do not re-verify `hostId` themselves, so the handshake check is load-bearing for every moderator event, not just the REST routes.
- State sent to non-moderators is sanitized; moderator-only data and duplicate Fast Money warnings are not broadcast to everyone.
- Acknowledgements use `{ ok: true, ... }`; failures use the `app_error` event with `ok`, `code`, `message`, and `statusCode`.
- Reconnect sends `sync_state`; lobby clients can receive `lobby_waiting` and player join notifications.
- Presence reports distinct online players per team, online player IDs, moderator connection, and audience count.
- PING/PONG latency tracking keeps a five-sample RTT moving average. Buzzer presses wait 150 ms, estimate one-way delay as RTT/2, deduplicate each player/team, and resolve all distinct presses in estimated click order. Client and server `performance.now()` values are not comparable and must not be used as a false cheat detector.

### Persistence queue

The current queue is deliberately in-process, not BullMQ:

- Live snapshots are latest-wins per session, coalesced to one Mongo write in flight, guarded by the live-state `version`.
- Event records buffer for 200 ms or 50 events and insert with unordered writes. Preassigned event IDs make retries idempotent for duplicate keys.
- Side tasks are coalesced by key and retried up to five times.
- Retry backoff starts at 500 ms and doubles; shutdown caps backoff at 250 ms.
- Shutdown drains event, state, and side-task work before Redis teardown. A hard failure may drop a snapshot/event after five attempts while Redis remains authoritative for live state.

## Migration Principles

1. Use a strangler migration with capability-level ownership. Do not run uncontrolled TypeScript and Go writes against the same live game.
2. Treat HTTP JSON, Mongo documents, Redis serialization/locking, Socket.IO events, and token claims as compatibility contracts.
3. Keep domain state transitions independent from HTTP and Socket.IO transport.
4. Prefer explicit Go structs and repositories over reproducing Mongoose APIs.
5. Make every ownership change reversible with a routing or feature-flag switch.

## Phased Execution

### Phase 0: Freeze and measure the contracts

- Generate an inventory from the actual route files and `SOCKET_EVENTS` constants.
- Add HTTP contract fixtures for every route, including public join/survey paths and host authorization failures.
- Capture representative Mongo documents, indexes, populated responses, soft-deleted records, and ObjectId serialization.
- Capture Redis state, cache keys, TTLs, lock contention, presence changes, and version progression.
- Add Socket.IO contract tests for handshake variants, room membership, state sanitization, acknowledgements, errors, reconnect sync, and every event group.
- Add a full-game fixture covering normal rounds, buzz ordering, face-off decisions, steals, disqualification, abort, Fast Money, and completion.

### Phase 1: Build the Go foundation

Create a Go service with `cmd/server` and internal packages for config, HTTP, auth, repositories, live game, realtime, AI, persistence, Mongo, Redis, and responses. Implement:

- Existing environment names and startup validation.
- JSON response/error envelopes and ObjectId-compatible IDs.
- CORS, cookies, body parsing, request IDs, logs, and graceful shutdown.
- Separate liveness and readiness checks for MongoDB and Redis.
- MongoDB and Redis clients with connection timeouts and shutdown handling.
- A deployment path that can run beside the TypeScript service.

Required environment names include `PORT`, `NODE_ENV`, `MONGO_URI`, `REDIS_URL`, the five JWT secret/expiry values (`JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `JWT_ACCESS_EXPIRES`, `JWT_REFRESH_EXPIRES_SHORT`, and `JWT_REFRESH_EXPIRES_LONG`), `MIN_QUESTIONS_TO_PUBLISH`, `LLM_PROVIDER`, `LLM_MODEL`, `LLM_API_KEY`, Fast Money time defaults, cookie settings, `ALLOWED_ORIGINS`, and refresh-token grace settings.

### Phase 2: Port data access and read-only APIs

Implement schema-compatible repositories for User, Game, GameSession, Survey, SurveyRespondent, Questions, GameLive, and GameLiveEvent. Preserve soft-delete filtering and populate projections. Port reads in this order:

1. Users and games
2. Surveys and public survey lookup
3. Sessions and public join/rejoin lookup
4. Questions and live-game state

Compare Go and TypeScript responses against the same sanitized database fixtures before routing traffic.

### Phase 3: Port authentication and session entry

Port registration, login, refresh, logout, password change, access-token middleware, host authorization, session creation, public player join/rejoin, and player-token generation. Verify Argon2, JWT claims/secrets/expiry, refresh-cookie attributes, token rotation and grace behavior, device identity, and player-token binding.

### Phase 4: Port CRUD writes

Port game, survey, respondent, question, user, session, team, and player writes. Test defaults, validation, status transitions, nested-array updates, soft delete/restore, transactions, duplicate codes, and timestamps. Keep public response submission and public session joining separate from authenticated management operations.

### Phase 5: Port the live-game domain

Implement table-driven state-machine tests before connecting sockets. Cover:

- Game start settings, question selection/splitting, multipliers, and round progression.
- Buzzer nominations, 150 ms buffered presses, face-off attempts, play/pass decisions, reopen/reorder, and turn rotation.
- Answer normalization/matching, reveal, strikes, steals, scoring, round end, random pick, disqualified teams, tie/results, abort, and completion.
- Fast Money player selection, five questions, tap/no-match/pass/reveal, duplicate handling, per-player time, pause/resume/time-up, phase transitions, target score, and completion.

All mutations must use the Redis lock and version rules. Test lock failure, stale versions, Redis seed races, buzzer contention, and recovery from Mongo snapshots.

### Phase 6: Port persistence and Redis compatibility

Implement the exact key builders, JSON shape, TTLs, `NX` seed, Lua release semantics, buzzer lock, presence counters/hash, state version guard, latest-wins snapshot queue, event batching/idempotency, side-task coalescing, retries, and shutdown drain. Validate with a real Redis integration test and two Go instances.

### Phase 7: Keep Socket.IO compatibility during cutover

Initially retain the TypeScript Socket.IO gateway if that reduces risk. It may forward authenticated commands to Go while continuing to emit the existing events. Then move handler ownership to Go only after the Go domain passes the full event contract suite.

Do not replace Socket.IO with raw WebSockets as part of the language migration. A raw WebSocket protocol is a separate client migration requiring a versioned envelope, request IDs, heartbeats, room semantics, reconnect sync, slow-client handling, and an explicit frontend rollout.

### Phase 8: Port AI behavior

Create a provider interface for OpenAI and Anthropic, preserving model, temperature, prompts, structured output requirements, parser behavior, timeouts, and error responses. The environment and dependency list mention Google Gemini, but the current `getLLM()` implementation does not select a Google provider; decide and record whether Google support is part of the target contract before implementing it in Go. Use recorded provider responses for parser tests and keep live model calls out of normal CI.

### Phase 9: Dual run, cut over, and remove TypeScript

1. Run Go in shadow mode for reads and compare response fixtures.
2. Route read-only capabilities independently.
3. Route authentication and public join/rejoin only after token compatibility passes.
4. Route CRUD writes after repository and transaction checks pass.
5. Run isolated live games with Go-owned Redis state and persistence.
6. Move Socket.IO command and broadcast ownership.
7. Run multi-instance soak and failure tests.
8. Keep a rapid route-level rollback to TypeScript until post-cutover persistence and gameplay checks pass.
9. Remove TypeScript capability by capability, then remove the old service only after rollback is no longer required.

## Testing and Acceptance Gates

- Go builds and starts with the documented environment.
- HTTP contract tests match paths, auth boundaries, status codes, envelopes, pagination, errors, cookies, and null/empty behavior.
- Ownership tests confirm a logged-in user who is not `Game.createdBy`/`GameSession.hostId` gets 403 on every owner/host-gated route, and a moderator socket handshake is rejected when the JWT's user is not `GameSession.hostId` for the target session.
- Existing Argon2 passwords, access tokens where required, refresh tokens, player tokens, and device rejoin flows work.
- Mongo repository tests cover all eight models, indexes, population, transactions, soft deletes, and real representative documents.
- Redis tests cover serialization, TTL, locks, stale versions, buzzer contention, presence, and two-instance adapter behavior.
- Socket tests cover every event constant, handshake, room, role, acknowledgement, sanitization, reconnect, and error contract.
- Domain tests cover complete normal and Fast Money games plus invalid transitions and concurrent buzzers.
- Persistence tests cover batching, coalescing, retries, duplicate event IDs, version guards, shutdown drain, and failure/drop logging.
- AI parser fixtures pass without live provider calls.
- Load and soak tests cover concurrent buzzers, scoring, reconnects, broadcasts, Mongo latency, Redis failure, and graceful shutdown.

## Known Baseline Gaps to Resolve Before Cutover

- `initSocket(server)` is called without awaiting its promise, so startup readiness is not currently synchronized with Redis adapter connection.
- `/health` reports MongoDB only even though Redis is required for socket/live operation.
- `latency.compensation.test.ts` still expects rejection based on client timestamps, while the implementation intentionally removed that invalid cross-process clock comparison. Update the test contract before using it as a migration gate.
- The declared Google Gemini dependency is not selected by the current LLM factory; provider scope must be decided explicitly.
- Socket latency tracking and pending buzzer state are process-local; Redis protects cross-instance buzzer ownership, but RTT histories and pending buffers are not shared. Test the chosen behavior during instance failover.
- Shutdown handlers exist in both the main server and socket module. Go should have one coordinated shutdown owner that closes HTTP, Socket.IO, queue, MongoDB, and Redis in a defined order.
- There is no documented automated HTTP, Mongo, Redis, Socket.IO, or end-to-end contract suite yet; Phase 0 must establish these fixtures before traffic migration.
- Resolved on `multiplayer`: game and session mutation routes, and the moderator socket handshake, previously accepted any authenticated user/any self-declared `role: "moderator"` client without checking `Game.createdBy`/`GameSession.hostId`. `assertGameOwner`, `assertSessionHost`, and the handshake's `GameSessionService.isHost` check now enforce this - see the route table and Authentication/Realtime sections above. Phase 0 fixtures and Phase 3/7 tests must include a non-owner/non-host case expecting 403 (REST) or a rejected handshake (socket).

## Rollback

- Route each capability back to TypeScript independently.
- Never let both implementations mutate one live session without a single lock/state owner.
- Preserve Redis state version and event ordering when changing ownership.
- Drain or invalidate queued persistence work before ownership changes.
- Back up MongoDB before Go takes write ownership.
- For a live-game rollback, stop Go commands first, verify the TypeScript process can read the current Redis state, then restore broadcast/command ownership.

## Definition of Done

The migration is complete when all contract and integration gates pass, existing clients can authenticate, join, reconnect, and complete games without an unplanned protocol change, MongoDB and Redis data remain compatible, multi-instance realtime behavior is verified, operational readiness is observable, and every capability has a tested rollback path.
