package fx

import (
	"go.uber.org/fx"

	"github.com/gowool/pages"
)

func SeederBoot(seeder pages.Seeder, lc fx.Lifecycle) {
	lc.Append(fx.StartHook(seeder.Boot))
}
