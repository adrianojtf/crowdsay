package poll

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	pollGroup := r.Group("/polls")
	{
		pollGroup.POST("/create", createPollHandler)
		pollGroup.POST("/:id/vote", votePollHandler)
		pollGroup.GET("/:id", getPollHandler)
		pollGroup.GET("/", getAllPollsHandler)
		pollGroup.GET("/:id/stats", getPollStatsHandler)
	}
}
func getPollStatsHandler(c *gin.Context) {
	pollIDStr := c.Param("id")
	pollID, err := strconv.ParseInt(pollIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid poll id",
			"next":  []string{"GET /polls"},
		})
		return
	}
	poll, err := GetPoll(pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"next":  []string{"GET /polls"},
		})
		return
	}
	total := 0
	for _, v := range poll.Results {
		total += v
	}
	percent := make(map[string]float64)
	for opt, v := range poll.Results {
		if total > 0 {
			percent[opt] = (float64(v) / float64(total)) * 100
		} else {
			percent[opt] = 0
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"poll_id":     poll.ID,
		"question":    poll.Question,
		"options":     poll.Options,
		"votes":       poll.Results,
		"percent":     percent,
		"total_votes": total,
		"next":        []string{"GET /polls/" + pollIDStr, "POST /polls/" + pollIDStr + "/vote"},
	})
}

func createPollHandler(c *gin.Context) {
	var p Poll
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"next":  []string{"POST /polls/create"},
		})
		return
	}
	if err := CreatePoll(&p); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
			"next":  []string{"POST /polls/create"},
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "poll created",
		"id":      p.ID,
		"next": []string{
			"GET /polls/" + strconv.FormatInt(p.ID, 10),
			"POST /polls/" + strconv.FormatInt(p.ID, 10) + "/vote",
			"GET /polls/" + strconv.FormatInt(p.ID, 10) + "/stats",
		},
	})
}

func votePollHandler(c *gin.Context) {
	pollIDStr := c.Param("id")
	pollID, err := strconv.ParseInt(pollIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid poll id",
			"next":  []string{"GET /polls"},
		})
		return
	}
	var req struct {
		Option string `json:"option"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"next":  []string{"GET /polls/" + pollIDStr},
		})
		return
	}
	ip := c.ClientIP()
	if err := VotePoll(pollID, req.Option, ip); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"next":  []string{"GET /polls/" + pollIDStr},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "vote registered",
		"next": []string{"GET /polls/" + pollIDStr,
			"GET /polls/" + pollIDStr + "/stats"},
	})
}

func getPollHandler(c *gin.Context) {
	pollIDStr := c.Param("id")
	pollID, err := strconv.ParseInt(pollIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid poll id",
			"next":  []string{"GET /polls"},
		})
		return
	}
	poll, err := GetPoll(pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"next":  []string{"GET /polls"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"poll": poll,
		"next": []string{"POST /polls/" + pollIDStr + "/vote"},
	})
}

func getAllPollsHandler(c *gin.Context) {
	polls, err := GetAllPolls()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"next":  []string{"POST /polls/create"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"polls": polls,
		"next":  []string{"POST /polls/create"},
	})
}
