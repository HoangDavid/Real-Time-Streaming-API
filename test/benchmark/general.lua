math.randomseed(os.time()) -- Seed for randomization

wrk.headers["Content-Type"] = "application/json"
wrk.method = "POST"

-- Tracks whether a streamId has been started
startedStreams = {}

-- Weighted operations after the stream has been started
operations = {
    { name = "send", weight = 6 }, -- Higher probability for sending messages
    { name = "end", weight = 1 }  -- Lower probability for ending streams
}

-- Calculate total weight for random selection
totalWeight = 0
for _, op in ipairs(operations) do
    totalWeight = totalWeight + op.weight
end

-- Function to select a random operation based on weight
function getRandomOperation()
    local rand = math.random(1, totalWeight)
    for _, op in ipairs(operations) do
        if rand <= op.weight then
            return op.name
        else
            rand = rand - op.weight
        end
    end
end

-- Function to generate a dynamic request
function request()
    -- Generate a unique `streamId` between 1 and 900
    local streamId = math.random(1, 800)

    -- Ensure `/start` is called first for each unique `streamId`
    if not startedStreams[streamId] then
        startedStreams[streamId] = true
        local path = "/stream/" .. streamId .. "/start"
        return wrk.format("POST", path, nil, nil) -- No payload for `/start`
    else
        -- After `/start`, select a random weighted operation
        local operation = getRandomOperation()
        local path = "/stream/" .. streamId .. "/" .. operation

        -- Generate appropriate payloads for each operation
        local payload = nil
        if operation == "send" then
            -- Randomly vary the message length for more realistic traffic
            local messageSize = math.random(10, 200)
            local message = string.rep("A", messageSize)
            payload = '{"data":"Hello Stream ' .. streamId .. ' - ' .. message .. '"}'
        elseif operation == "end" then
            payload = nil -- No payload for `/end`
        end

        return wrk.format(nil, path, nil, payload)
    end
end
