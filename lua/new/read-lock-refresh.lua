-- Lock timer refresh script.

-- This script checks if the lock was acquired by the same owner and then refreshes the expiration timer.

-- KEYS = []GLOB_LOCK_KEY
-- ARGV = [TOKEN, EXPIRATION_TIMEOUT]

-- check that global lock is ours

for _, v in pairs(KEYS) do
    if redis.call("GET", v) == ARGV[1] then
        -- update global lock timeout
        redis.call("PEXPIRE", v, ARGV[2])
    end
end
return 1

