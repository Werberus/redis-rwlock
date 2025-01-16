package rwlock

import (
	"github.com/redis/go-redis/v9"
)

var (
	acquireReadLock = redis.NewScript(readLockScript)
	releaseReadLock = redis.NewScript(readUnlockScript)

	refreshLock      = redis.NewScript(lockRefreshScript)
	refreshReadLock  = redis.NewScript(lockReadRefreshScript)
	acquireWriteLock = redis.NewScript(writeLockScript)
	releaseWriteLock = redis.NewScript(writeUnlockScript)
)

// KEYS = []GLOB_LOCK_KEY
// ARGV = [TOKEN, EXPIRATION_TIMEOUT, WRITER_PREFERRING, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]
const readLockScript = `
local function lock(arrKeys, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end
    local key = arrKeys[i]
    if ARGV[3] ~= 0 and redis.call("EXISTS", ARGV[5] .. key) == 1 then
        return 0
    else
        if redis.call("INCR", ARGV[4] .. key) > 0  then
            if redis.call("SET", key, ARGV[1], "PX", ARGV[2], "NX") then
                if lock(arrKeys, i) == 1 then
                    return 1
                end
            else
                if redis.call("GET", key) == ARGV[1] then
                    redis.call("PEXPIRE", key, ARGV[2])
                    if lock(arrKeys, i) == 1 then
                       return 1
                    end
                else
                    redis.call("DECR", ARGV[4] .. key)
                    return 0
                end
            end
        else
            if lock(arrKeys, i) == 1 then
                return 1
            end
        end
        redis.call("DECR", ARGV[4] .. key)
        return 0
    end
end
return lock(KEYS, 0)`

// KEYS = []GLOB_LOCK_KEY
// ARGV = [TOKEN, READ_LOCK_REF_COUNT]
const readUnlockScript = `
local function unlock(arrKeys, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end
    if redis.call("DECR", ARGV[2] .. arrKeys[i]) == 0 then
        if redis.call("GET", arrKeys[i]) == ARGV[1] then
            redis.call("DEL", arrKeys[i], ARGV[2] .. arrKeys[i])
            return unlock(arrKeys, i)
        else
           unlock(arrKeys, i)
           return 0
        end
    else
       return unlock(arrKeys, i)
    end
end
return unlock(KEYS, 0)`

// KEYS = []GLOB_LOCK_KEY
// ARGV = [TOKEN, EXPIRATION_TIMEOUT]
const lockReadRefreshScript = `
for _, v in pairs(KEYS) do
    if redis.call("GET", v) == ARGV[1] then
        redis.call("PEXPIRE", v, ARGV[2])
    end
end
return 1`

// KEYS = [GLOB_LOCK_KEY]
// ARGV = [TOKEN, EXPIRATION_TIMEOUT]
const lockRefreshScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
    return 0
end`

// KEYS = []GLOB_LOCK_KEY
// ARGV = [TOKEN, EXPIRATION_TIMEOUT, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]
const writeLockScript = `
redis.call("SET", ARGV[4]..KEYS[1], 1, "PX", ARGV[2])
if redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2], "NX") then
    redis.call("DEL", ARGV[3]..KEYS[1], ARGV[4]..KEYS[1])
    return 1
else
    return 0
end`

// KEYS = [GLOB_LOCK_KEY]
// ARGV = [TOKEN]
const writeUnlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("DEL", KEYS[1])
    return 1
else
    return 0
end`
