package service

// Weights are fractions of the full score. Both profiles sum to 1;
// interpolation therefore preserves the 0..100 scale.
type impactWeights struct {
	KDA, Assists, HeroDamage, TowerDamage float64
	Healing, Stuns, Wards, Participation  float64
	GPM, XPM, NetWorth, Survival, Win     float64
}

// Return values rather than mutable package-level configuration.
func defaultImpactWeights() (core, support impactWeights) {
	core = impactWeights{
		KDA: .17, Assists: .04, HeroDamage: .22, TowerDamage: .14,
		Healing: .01, Stuns: .02, Wards: 0, Participation: .06,
		GPM: .10, XPM: .07, NetWorth: .08, Survival: .04, Win: .05,
	}
	support = impactWeights{
		KDA: .10, Assists: .13, HeroDamage: .08, TowerDamage: .03,
		Healing: .10, Stuns: .14, Wards: .12, Participation: .14,
		GPM: .02, XPM: .03, NetWorth: .02, Survival: .04, Win: .05,
	}
	return core, support
}
