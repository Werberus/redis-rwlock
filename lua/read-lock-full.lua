-- Reader lock acquire script.

-- If key WRITER_PREFERRING is not 0 this script retreats if writer has set it's intention to acquire lock.
-- The script increments number of shared locks and if it was the first to do that it tries to acquire lock.
-- If it fails to acquire lock then it decrements number of shared locks back to previous value.

-- KEYS = [GLOB_LOCK_KEY, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]
-- ARGV = [TOKEN, EXPIRATION_TIMEOUT, WRITER_PREFERRING]

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

return lock(arrKeys, "", 0)



