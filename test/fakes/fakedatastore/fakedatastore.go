package fakedatastore

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"github.com/spiffe/spire/pkg/common/util"
	"github.com/spiffe/spire/pkg/server/datastore"
	sql "github.com/spiffe/spire/pkg/server/datastore/sqlstore"
	"github.com/spiffe/spire/proto/spire/common"
)

var (
	ctx = context.Background()
)

type DataStore struct {
	ds   datastore.DataStore
	errs []error
}

var _ datastore.DataStore = (*DataStore)(nil)

func New(tb testing.TB) *DataStore {
	log, _ := test.NewNullLogger()

	ds := sql.New(log)
	ds.SetUseServerTimestamps(true)

	tmpDir := tb.TempDir()
	dbPath := filepath.Join(tmpDir, "spire.db")
	dbPath = url.PathEscape(dbPath)

	err := ds.Configure(ctx, fmt.Sprintf(`
		database_type = "sqlite3"
		connection_string = "file:%s"
	`, dbPath))
	require.NoError(tb, err)

	tb.Cleanup(func() {
		ds.Close()
	})

	return &DataStore{
		ds: ds,
	}
}

func (s *DataStore) CreateBundle(ctx context.Context, bundle *common.Bundle) (*common.Bundle, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.CreateBundle(ctx, bundle)
}

func (s *DataStore) UpdateBundle(ctx context.Context, bundle *common.Bundle, mask *common.BundleMask) (*common.Bundle, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.UpdateBundle(ctx, bundle, mask)
}

func (s *DataStore) SetBundle(ctx context.Context, bundle *common.Bundle) (*common.Bundle, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.SetBundle(ctx, bundle)
}

func (s *DataStore) AppendBundle(ctx context.Context, bundle *common.Bundle) (*common.Bundle, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.AppendBundle(ctx, bundle)
}

func (s *DataStore) CountBundles(ctx context.Context) (int32, error) {
	if err := s.getNextError(); err != nil {
		return 0, err
	}

	return s.ds.CountBundles(ctx)
}

func (s *DataStore) DeleteBundle(ctx context.Context, trustDomain string, mode datastore.DeleteMode) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.DeleteBundle(ctx, trustDomain, mode)
}

func (s *DataStore) FetchBundle(ctx context.Context, trustDomain string) (*common.Bundle, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchBundle(ctx, trustDomain)
}

func (s *DataStore) ListBundles(ctx context.Context, req *datastore.ListBundlesRequest) (*datastore.ListBundlesResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	resp, err := s.ds.ListBundles(ctx, req)
	if err == nil {
		// Sorting helps unit-tests have deterministic assertions.
		sort.Slice(resp.Bundles, func(i, j int) bool {
			return resp.Bundles[i].TrustDomainId < resp.Bundles[j].TrustDomainId
		})
	}
	return resp, err
}

func (s *DataStore) PruneBundle(ctx context.Context, trustDomainID string, expiresBefore time.Time) (bool, error) {
	if err := s.getNextError(); err != nil {
		return false, err
	}
	return s.ds.PruneBundle(ctx, trustDomainID, expiresBefore)
}

func (s *DataStore) CountAttestedNodes(ctx context.Context, req *datastore.CountAttestedNodesRequest) (int32, error) {
	if err := s.getNextError(); err != nil {
		return 0, err
	}
	return s.ds.CountAttestedNodes(ctx, req)
}

func (s *DataStore) CreateAttestedNode(ctx context.Context, node *common.AttestedNode) (*common.AttestedNode, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.CreateAttestedNode(ctx, node)
}

func (s *DataStore) FetchAttestedNode(ctx context.Context, spiffeID string) (*common.AttestedNode, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchAttestedNode(ctx, spiffeID)
}

func (s *DataStore) ListAttestedNodes(ctx context.Context, req *datastore.ListAttestedNodesRequest) (*datastore.ListAttestedNodesResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListAttestedNodes(ctx, req)
}

func (s *DataStore) UpdateAttestedNode(ctx context.Context, node *common.AttestedNode, mask *common.AttestedNodeMask) (*common.AttestedNode, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.UpdateAttestedNode(ctx, node, mask)
}

func (s *DataStore) DeleteAttestedNode(ctx context.Context, spiffeID string) (*common.AttestedNode, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.DeleteAttestedNode(ctx, spiffeID)
}

func (s *DataStore) PruneAttestedExpiredNodes(ctx context.Context, expiredBefore time.Time, includeNonReattestable bool) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneAttestedExpiredNodes(ctx, expiredBefore, includeNonReattestable)
}

func (s *DataStore) ListAttestedNodeEvents(ctx context.Context, req *datastore.ListAttestedNodeEventsRequest) (*datastore.ListAttestedNodeEventsResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListAttestedNodeEvents(ctx, req)
}

func (s *DataStore) PruneAttestedNodeEvents(ctx context.Context, olderThan time.Duration) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneAttestedNodeEvents(ctx, olderThan)
}

func (s *DataStore) CreateAttestedNodeEventForTesting(ctx context.Context, event *datastore.AttestedNodeEvent) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.CreateAttestedNodeEventForTesting(ctx, event)
}

func (s *DataStore) DeleteAttestedNodeEventForTesting(ctx context.Context, eventID uint) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.DeleteAttestedNodeEventForTesting(ctx, eventID)
}

func (s *DataStore) FetchAttestedNodeEvent(ctx context.Context, eventID uint) (*datastore.AttestedNodeEvent, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchAttestedNodeEvent(ctx, eventID)
}

func (s *DataStore) TaintX509CA(ctx context.Context, trustDomainID string, subjectKeyIDToTaint string) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.TaintX509CA(ctx, trustDomainID, subjectKeyIDToTaint)
}

func (s *DataStore) RevokeX509CA(ctx context.Context, trustDomainID string, subjectKeyIDToRevoke string) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.RevokeX509CA(ctx, trustDomainID, subjectKeyIDToRevoke)
}

func (s *DataStore) TaintJWTKey(ctx context.Context, trustDomainID string, authorityID string) (*common.PublicKey, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.TaintJWTKey(ctx, trustDomainID, authorityID)
}

func (s *DataStore) RevokeJWTKey(ctx context.Context, trustDomainID string, authorityID string) (*common.PublicKey, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.RevokeJWTKey(ctx, trustDomainID, authorityID)
}

func (s *DataStore) SetNodeSelectors(ctx context.Context, spiffeID string, selectors []*common.Selector) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.SetNodeSelectors(ctx, spiffeID, selectors)
}

func (s *DataStore) ListNodeSelectors(ctx context.Context, req *datastore.ListNodeSelectorsRequest) (*datastore.ListNodeSelectorsResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListNodeSelectors(ctx, req)
}

func (s *DataStore) GetNodeSelectors(ctx context.Context, spiffeID string, dataConsistency datastore.DataConsistency) ([]*common.Selector, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	selectors, err := s.ds.GetNodeSelectors(ctx, spiffeID, dataConsistency)
	if err == nil {
		// Sorting helps unit-tests have deterministic assertions.
		util.SortSelectors(selectors)
	}
	return selectors, err
}

func (s *DataStore) CountRegistrationEntries(ctx context.Context, req *datastore.CountRegistrationEntriesRequest) (int32, error) {
	if err := s.getNextError(); err != nil {
		return 0, err
	}
	return s.ds.CountRegistrationEntries(ctx, req)
}

func (s *DataStore) CreateRegistrationEntry(ctx context.Context, entry *common.RegistrationEntry) (*common.RegistrationEntry, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.CreateRegistrationEntry(ctx, entry)
}

func (s *DataStore) CreateOrReturnRegistrationEntry(ctx context.Context, entry *common.RegistrationEntry) (*common.RegistrationEntry, bool, error) {
	if err := s.getNextError(); err != nil {
		return nil, false, err
	}
	return s.ds.CreateOrReturnRegistrationEntry(ctx, entry)
}

func (s *DataStore) FetchRegistrationEntry(ctx context.Context, entryID string) (*common.RegistrationEntry, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchRegistrationEntry(ctx, entryID)
}

func (s *DataStore) FetchRegistrationEntries(ctx context.Context, entryIDs []string) (map[string]*common.RegistrationEntry, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchRegistrationEntries(ctx, entryIDs)
}

func (s *DataStore) ListRegistrationEntries(ctx context.Context, req *datastore.ListRegistrationEntriesRequest) (*datastore.ListRegistrationEntriesResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	resp, err := s.ds.ListRegistrationEntries(ctx, req)
	if err == nil {
		// Sorting helps unit-tests have deterministic assertions.
		util.SortRegistrationEntries(resp.Entries)
	}
	return resp, err
}

func (s *DataStore) UpdateRegistrationEntry(ctx context.Context, entry *common.RegistrationEntry, mask *common.RegistrationEntryMask) (*common.RegistrationEntry, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.UpdateRegistrationEntry(ctx, entry, mask)
}

func (s *DataStore) DeleteRegistrationEntry(ctx context.Context, entryID string) (*common.RegistrationEntry, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.DeleteRegistrationEntry(ctx, entryID)
}

func (s *DataStore) PruneRegistrationEntries(ctx context.Context, expiresBefore time.Time) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneRegistrationEntries(ctx, expiresBefore)
}

func (s *DataStore) ListRegistrationEntryEvents(ctx context.Context, req *datastore.ListRegistrationEntryEventsRequest) (*datastore.ListRegistrationEntryEventsResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListRegistrationEntryEvents(ctx, req)
}

func (s *DataStore) PruneRegistrationEntryEvents(ctx context.Context, olderThan time.Duration) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneRegistrationEntryEvents(ctx, olderThan)
}

func (s *DataStore) CreateRegistrationEntryEventForTesting(ctx context.Context, event *datastore.RegistrationEntryEvent) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.CreateRegistrationEntryEventForTesting(ctx, event)
}

func (s *DataStore) DeleteRegistrationEntryEventForTesting(ctx context.Context, eventID uint) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.DeleteRegistrationEntryEventForTesting(ctx, eventID)
}

func (s *DataStore) FetchRegistrationEntryEvent(ctx context.Context, eventID uint) (*datastore.RegistrationEntryEvent, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchRegistrationEntryEvent(ctx, eventID)
}

func (s *DataStore) CreateJoinToken(ctx context.Context, token *datastore.JoinToken) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.CreateJoinToken(ctx, token)
}

func (s *DataStore) FetchJoinToken(ctx context.Context, token string) (*datastore.JoinToken, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchJoinToken(ctx, token)
}

func (s *DataStore) DeleteJoinToken(ctx context.Context, token string) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.DeleteJoinToken(ctx, token)
}

func (s *DataStore) PruneJoinTokens(ctx context.Context, expiresBefore time.Time) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneJoinTokens(ctx, expiresBefore)
}

func (s *DataStore) CreateFederationRelationship(c context.Context, fr *datastore.FederationRelationship) (*datastore.FederationRelationship, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.CreateFederationRelationship(c, fr)
}

func (s *DataStore) DeleteFederationRelationship(c context.Context, trustDomain spiffeid.TrustDomain) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.DeleteFederationRelationship(c, trustDomain)
}

func (s *DataStore) FetchFederationRelationship(c context.Context, trustDomain spiffeid.TrustDomain) (*datastore.FederationRelationship, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchFederationRelationship(c, trustDomain)
}

func (s *DataStore) ListFederationRelationships(ctx context.Context, req *datastore.ListFederationRelationshipsRequest) (*datastore.ListFederationRelationshipsResponse, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListFederationRelationships(ctx, req)
}

func (s *DataStore) UpdateFederationRelationship(ctx context.Context, fr *datastore.FederationRelationship, mask *types.FederationRelationshipMask) (*datastore.FederationRelationship, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.UpdateFederationRelationship(ctx, fr, mask)
}

func (s *DataStore) FetchCAJournal(ctx context.Context, activeX509AuthorityID string) (*datastore.CAJournal, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.FetchCAJournal(ctx, activeX509AuthorityID)
}

func (s *DataStore) ListCAJournalsForTesting(ctx context.Context) ([]*datastore.CAJournal, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.ListCAJournalsForTesting(ctx)
}

func (s *DataStore) SetCAJournal(ctx context.Context, caJournal *datastore.CAJournal) (*datastore.CAJournal, error) {
	if err := s.getNextError(); err != nil {
		return nil, err
	}
	return s.ds.SetCAJournal(ctx, caJournal)
}

func (s *DataStore) PruneCAJournals(ctx context.Context, allCAsExpireBefore int64) error {
	if err := s.getNextError(); err != nil {
		return err
	}
	return s.ds.PruneCAJournals(ctx, allCAsExpireBefore)
}

func (s *DataStore) SetNextError(err error) {
	s.errs = []error{err}
}

func (s *DataStore) AppendNextError(err error) {
	s.errs = append(s.errs, err)
}

func (s *DataStore) getNextError() error {
	if len(s.errs) == 0 {
		return nil
	}
	err := s.errs[0]
	s.errs = s.errs[1:]
	return err
}
