-- phone_code:login:1xxxxxxxxxx
local key = KEYS[1]
-- 验证次数,设置重复三次
-- phone_code:login:152xxxxxxxx:cnt
local cntKey = key..":cnt"
-- 验证码
local val = ARGV[1]
--  获取过期时间
local ttl = tonumber(redis.call("ttl",key))
if ttl == -1 then
    -- key 存在，但是无过期时间
    return -2
    -- 10 分钟有效期，每一分钟只能一次
    -- 540 = 600 -60
    -- 不存在key ttl 的结果为 -2
elseif ttl == -2 or ttl < 540  then
    redis.call("set",key,val)
    redis.call("expire",key,600)
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
    -- 成功
    return 0
else
    -- 发送太频繁
    return -1
end