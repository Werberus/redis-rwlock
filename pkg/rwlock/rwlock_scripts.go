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

// KEYS = [GLOB_LOCK_KEY, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]
// ARGV = [TOKEN, EXPIRATION_TIMEOUT, WRITER_PREFERRING]
const readLockScript = `
local function mSplit(inputStr, sep)
    if sep == nil then
        sep = "%s"
    end
    local t = {}
    for str in string.gmatch(inputStr, "([^"..sep.."]+)") do
        table.insert(t, str)
    end
    return t
end

local function lock(arrKeys, key, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end

    if key == "" then
        key = arrKeys[i]
    else
        key=key.."/"..arrKeys[i]
    end

    if ARGV[3] ~= 0 and redis.call("EXISTS", KEYS[3] .. key) == 1 then
        -- failed
        return 0
    else
        -- increment ref counter. if first, acquire global lock.
        if redis.call("INCR", KEYS[2] .. key) > 0  then
            -- acquire global lock
            if redis.call("SET", key, ARGV[1], "PX", ARGV[2], "NX") then
                -- global lock acquired. success
                if lock(arrKeys, key, i) == 1 then
                    return 1
                end
            else
                if redis.call("GET", key) == ARGV[1] then
					redis.call("PEXPIRE", key, ARGV[2])
                    -- global lock acquired. success
                    if lock(arrKeys, key, i) == 1 then
                       return 1
                    end
                else
                    -- global lock not acquired. decrement ref counter
                    redis.call("DECR", KEYS[2] .. key)
                    -- failed
                    return 0
                end
            end
        else
            -- global lock must be acquired by some other reader. success
            if lock(arrKeys, key, i) == 1 then
                return 1
            end
        end
        redis.call("DECR", KEYS[2] .. key)
        -- failed
        return 0
    end
end


local arrKeys = mSplit(KEYS[1], "/")

return lock(arrKeys, "", 0)`

// KEYS = [GLOB_LOCK_KEY, READ_LOCK_REF_COUNT]
// ARGV = [TOKEN]
const readUnlockScript = `
local function mSplit(inputstr, sep)
    if sep == nil then
        sep = "%s"
    end
    local t = {}
    for str in string.gmatch(inputstr, "([^"..sep.."]+)") do
        table.insert(t, str)
    end
    return t
end

local function unlock(arrKeys, key, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end

    if key == "" then
        key = arrKeys[i]
    else
        key=key.."/"..arrKeys[i]
    end

    if redis.call("DECR", KEYS[2] .. key) == 0 then
        -- check that global lock is ours
        if redis.call("GET", key) == ARGV[1] then
            -- release global lock
            redis.call("DEL", key, KEYS[2] .. key)
            -- success
            return unlock(arrKeys, key, i)
        else
            -- failed to release global lock
           return unlock(arrKeys, key, i)
        end
    else
        -- success
       return unlock(arrKeys, key, i)
    end
end

local arrKeys = mSplit(KEYS[1], "/")

return unlock(arrKeys, "", 0)`

// KEYS = [GLOB_LOCK_KEY]
// ARGV = [TOKEN, EXPIRATION_TIMEOUT]
const lockRefreshScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    -- update global lock timeout
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
    -- failed to update global lock timeout
    return 0
end`

const lockReadRefreshScript = `
local function mSplit(inputStr, sep)
    if sep == nil then
        sep = "%s"
    end
    local t = {}
    for str in string.gmatch(inputStr, "([^"..sep.."]+)") do
        table.insert(t, str)
    end
    return t
end


local function refresh(lockKey)
    if redis.call("GET", lockKey) == ARGV[1] then
        -- update global lock timeout
        return redis.call("PEXPIRE", lockKey, ARGV[2])
    else
        -- failed to update global lock timeout
        return 0
    end
end

local key1 = KEYS[1]
local arrKeys = mSplit(key1, "/")
for i = #arrKeys, 1, -1 do repeat
    if refresh(key1) then
        key1 = string.gsub(key1, "/".. arrKeys[i] .."/?$", "")
        break
    else
        return 0
    end
until true end
return 1`

// KEYS = [GLOB_LOCK_KEY, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]
// ARGV = [TOKEN, EXPIRATION_TIMEOUT]
const writeLockScript = `
redis.call("SET", KEYS[3]..KEYS[1], 1, "PX", ARGV[2])
if redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2], "NX") then
    redis.call("DEL", KEYS[2]..KEYS[1], KEYS[3]..KEYS[1])
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
