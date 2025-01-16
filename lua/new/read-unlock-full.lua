-- Reader lock release script.

-- This script decrements number of shared locks and if it was the last tries to release global lock.
-- If it fails to release global lock number of shared lock remains decremented.

-- KEYS = []GLOB_LOCK_KEY
-- ARGV = [TOKEN, READ_LOCK_REF_COUNT]

local function unlock(arrKeys, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end

    if redis.call("DECR", ARGV[2] .. arrKeys[i]) == 0 then
        -- check that global lock is ours
        if redis.call("GET", arrKeys[i]) == ARGV[1] then
            -- release global lock
            redis.call("DEL", arrKeys[i], ARGV[2] .. arrKeys[i])
            -- success
            return unlock(arrKeys, i)
        else
            -- failed to release global lock
           unlock(arrKeys, i)
           return 0
        end
    else
        -- success
       return unlock(arrKeys, i)
    end
end

return unlock(KEYS, 0)
