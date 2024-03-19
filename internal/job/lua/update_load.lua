local key = KEYS[1]
local loadKey = key..":load"
local load = ARGV[1]
local node = ARGV[2]
local expireTime = ARGV[3]

local exist = redis.call("EXISTS", key)

if exist == 1 then
    redis.call("set", key, node, "EX", expireTime)
    redis.call("set", loadKey, load, "EX", expireTime)
    return 1
end

local curLoad = tonumber(redis.call("get", loadKey))
local curNode = redis.call("get", key)

if curNode == node then
    redis.call("set", loadKey, load, "EX", expireTime)
    return 1
else
    if load < curLoad then
        redis.call("set", key, node, "EX", expireTime)
        redis.call("set", loadKey, load, "EX", expireTime)
        return 1
    else
        return -1
    end
end