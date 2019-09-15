local cursor = "0"
local keys = {}
repeat
  local res = redis.call('scan', cursor, 'MATCH', ARGV[1], 'COUNT', 50)
  cursor = res[1]
  for i, v in ipairs(res[2]) do
    table.insert(keys, v)
  end
until( cursor == "0" )

if table.getn(keys) == 0 then
  return keys
else
  return redis.call('mget', unpack(keys))
end