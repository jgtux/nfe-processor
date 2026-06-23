package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nfe-processor/backend/internal/service"
)

type Handler struct {
	svc *service.NFeService
}

func New(svc *service.NFeService) *Handler {
	return &Handler{svc: svc}
}

type uploadFileResult struct {
	File  string `json:"file"`
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type uploadResponse struct {
	Message string             `json:"message"`
	Results []uploadFileResult `json:"results"`
}

// UploadXML godoc
// @Summary      Upload NF-e XML files
// @Description  Receives one or more NF-e XML files, validates and enqueues for async processing via RabbitMQ
// @Tags         nfe
// @Accept       multipart/form-data
// @Produce      json
// @Param        files  formData  file  true  "XML file(s)"
// @Success      202    {object}  Response[uploadResponse]
// @Failure      400    {object}  Response[any]
// @Router       /xml/upload [post]
func (h *Handler) UploadXML(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		Fail(c, http.StatusBadRequest, fmt.Errorf("invalid form data"))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		Fail(c, http.StatusBadRequest, fmt.Errorf("no files provided (field: files)"))
		return
	}

	var results []uploadFileResult

	for _, fh := range files {
		name := fh.Filename
		if len(name) < 4 || name[len(name)-4:] != ".xml" {
			ct := fh.Header.Get("Content-Type")
			if ct != "text/xml" && ct != "application/xml" {
				results = append(results, uploadFileResult{File: name, Error: "not an XML file"})
				continue
			}
		}

		f, err := fh.Open()
		if err != nil {
			results = append(results, uploadFileResult{File: fh.Filename, Error: "failed to open file"})
			continue
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			results = append(results, uploadFileResult{File: fh.Filename, Error: "failed to read file"})
			continue
		}

		id, err := h.svc.EnqueueXML(data)
		if err != nil {
			log.Printf("[handler] enqueue error file=%s: %v", fh.Filename, err)
			results = append(results, uploadFileResult{File: fh.Filename, Error: err.Error()})
			continue
		}
		results = append(results, uploadFileResult{File: fh.Filename, ID: id})
	}

	Created(c, uploadResponse{
		Message: "files received and enqueued for processing",
		Results: results,
	})
}

// ListNFes godoc
// @Summary      List all processed NF-es
// @Description  Returns all successfully processed NF-es, excluding quarantined (error) records
// @Tags         nfe
// @Produce      json
// @Success      200  {object}  Response[[]domain.NFe]
// @Router       /nfe [get]
func (h *Handler) ListNFes(c *gin.Context) {
	list, err := h.svc.ListAll()
	if err != nil {
		Fail(c, http.StatusInternalServerError, err)
		return
	}
	OKList(c, list)
}

// ListUnidentified godoc
// @Summary      List unidentified NF-es
// @Description  Returns processed NF-es that could not be linked to any internal client
// @Tags         nfe
// @Produce      json
// @Success      200  {object}  Response[[]domain.NFe]
// @Router       /nfe/unidentified [get]
func (h *Handler) ListUnidentified(c *gin.Context) {
	list, err := h.svc.ListUnidentified()
	if err != nil {
		Fail(c, http.StatusInternalServerError, err)
		return
	}
	OKList(c, list)
}

// Quarantine godoc
// @Summary      List quarantined NF-es
// @Description  Returns NF-es that failed XSD validation, mod 11 checks, or parsing. Automatically deleted after QUARANTINE_TTL_DAYS days.
// @Tags         nfe
// @Produce      json
// @Success      200  {object}  Response[[]domain.NFe]
// @Router       /nfe/quarantine [get]
func (h *Handler) Quarantine(c *gin.Context) {
	list, err := h.svc.ListQuarantine()
	if err != nil {
		Fail(c, http.StatusInternalServerError, err)
		return
	}
	OKList(c, list)
}

// ClientSummary godoc
// @Summary      Client summary
// @Description  Returns purchase and sale counts grouped by internal client
// @Tags         nfe
// @Produce      json
// @Success      200  {object}  Response[[]domain.ClientSummary]
// @Router       /nfe/summary [get]
func (h *Handler) ClientSummary(c *gin.Context) {
	summary, err := h.svc.ClientSummary()
	if err != nil {
		Fail(c, http.StatusInternalServerError, err)
		return
	}
	OKList(c, summary)
}

// ListClients godoc
// @Summary      List internal clients
// @Description  Returns all clients registered in the internal clients mock API
// @Tags         clients
// @Produce      json
// @Success      200  {object}  Response[[]domain.InternalClient]
// @Router       /clients [get]
func (h *Handler) ListClients(c *gin.Context) {
	clients, err := h.svc.InternalClients()
	if err != nil {
		Fail(c, http.StatusInternalServerError, err)
		return
	}
	OKList(c, clients)
}

// Health godoc
// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200  {object}  Response[any]
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	OK(c, gin.H{"status": "ok"})
}
