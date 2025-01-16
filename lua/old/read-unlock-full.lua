-- Reader lock release script.

-- This script decrements number of shared locks and if it was the last tries to release global lock.
-- If it fails to release global lock number of shared lock remains decremented.

-- KEYS = [GLOB_LOCK_KEY, READ_LOCK_REF_COUNT]
-- ARGV = [TOKEN]

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

return unlock(arrKeys, "", 0)
