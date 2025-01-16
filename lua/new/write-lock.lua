-- Writer lock acquire script.

-- This script sets writer intention to acquire the lock.
-- Then in tries to acquire the lock. If lock is acquired successfully intention is reset.
-- If it failed to acquire the lock the intention remains set to prevent adding new writer-preferring readers.

-- KEYS = []GLOB_LOCK_KEY
-- ARGV = [TOKEN, EXPIRATION_TIMEOUT, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]

-- set writer intention to acquire global lock
redis.call("SET", ARGV[4]..KEYS[1], 1, "PX", ARGV[2])
-- acquire global lock
if redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2], "NX") then
    -- global lock acquired. reset intention. remove dangling readers refs.
    redis.call("DEL", ARGV[3]..KEYS[1], ARGV[4]..KEYS[1])
    -- success
    return 1
else
    -- failed to acquire global lock
    return 0
end
