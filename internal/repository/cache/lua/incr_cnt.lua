local key = KEY[1]
local cntKey = ARGV[1]
local delta = ARGV[2]
local exist = redis.call("EXISTS", key)

if exist == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    return 1
else
    return 0
end