-- Reader lock acquire script.

-- If key WRITER_PREFERRING is not 0 this script retreats if writer has set it's intention to acquire lock.
-- The script increments number of shared locks and if it was the first to do that it tries to acquire lock.
-- If it fails to acquire lock then it decrements number of shared locks back to previous value.

-- KEYS = []GLOB_LOCK_KEY
-- ARGV = [TOKEN, EXPIRATION_TIMEOUT, WRITER_PREFERRING, READ_LOCK_REF_COUNT, WRITER_LOCK_INTENT]

local function lock(arrKeys, i)
    i=i+1
    if i > #arrKeys then
        return 1
    end

    local key = arrKeys[i]

    if ARGV[3] ~= 0 and redis.call("EXISTS", ARGV[5] .. key) == 1 then
        -- failed
        return 0
    else
        -- increment ref counter. if first, acquire global lock.
        if redis.call("INCR", ARGV[4] .. key) > 0  then
            -- acquire global lock
            if redis.call("SET", key, ARGV[1], "PX", ARGV[2], "NX") then
                -- global lock acquired. success
                if lock(arrKeys, i) == 1 then
                    return 1
                end
            else
                if redis.call("GET", key) == ARGV[1] then
                    redis.call("PEXPIRE", key, ARGV[2])
                    -- global lock acquired. success
                    if lock(arrKeys, i) == 1 then
                       return 1
                    end
                else
                    -- global lock not acquired. decrement ref counter
                    redis.call("DECR", ARGV[4] .. key)
                    -- failed
                    return 0
                end
            end
        else
            -- global lock must be acquired by some other reader. success
            if lock(arrKeys, i) == 1 then
                return 1
            end
        end
        redis.call("DECR", ARGV[4] .. key)
        -- failed
        return 0
    end
end

return lock(KEYS, 0)



