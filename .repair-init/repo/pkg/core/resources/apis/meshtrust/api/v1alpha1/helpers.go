package v1alpha1

func GetAllTrustDomains(meshTrusts MeshTrustResourceList) map[string]struct{} {
	trustDomains := map[string]struct{}{}
	for _, meshTrust := range meshTrusts.Items {
		trustDomains[meshTrust.Spec.TrustDomain] = struct{}{}
	}
	return trustDomains
}
