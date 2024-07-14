-- phone_code:login:1xxxxxxxxxx
local key = KEYS[1]
-- 验证次数
-- phone_code:login:152xxxxxxxx:cnt
local cntKey = key..":cnt"
-- 输入的验证码
local expectedCode = ARGV[1]
-- 存入缓存的验证码
local code = redis.call("get",key)
-- 存入缓存的验证次数
local cnt = tonumber(redis.call("get",cntKey))
if cnt == nil or cnt <=0 then
    -- 验证次数已经使用完
    return -1
elseif expectedCode == code then
    -- 验证码正确
    -- 验证码作废
    redis.call("set",cntKey,-1)
    return 0
else
    -- 验证码输入错误
    redis.call("decr",cntKey)
    return -2
end