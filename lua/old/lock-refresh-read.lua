-- Lock timer refresh script.

-- This script checks if the lock was acquired by the same owner and then refreshes the expiration timer.

-- KEYS = [GLOB_LOCK_KEY]
-- ARGV = [TOKEN, EXPIRATION_TIMEOUT]

-- check that global lock is ours

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
return 1

