import{O as _,bU as h,c5 as p,k as m,cc as l,o,c as r,a,ch as b,w as i,b as e,e as t,bV as w,j as d,bX as y,bY as f}from"./index.bd548025.js";const k={name:"EnvironmentSwitcher",components:{KButton:_,KCard:h},data(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:{...p({environment:"config/getEnvironment"}),instructionsCtaText(){return this.environment==="universal"?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute(){return this.environment==="kubernetes"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}}},u=n=>(y("data-v-6ac9df0a"),n=n(),f(),n),g={class:"wizard-switcher"},S={class:"capitalize"},K={key:0},z={key:0},R=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),U={class:"text-center"},C=u(()=>t("br",null,null,-1)),B={key:1},E=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),V={class:"text-center"},I={key:1},N={key:0},W=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),O={class:"text-center"},T={key:1},j=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),x={class:"text-center"};function D(n,G,X,Y,s,q){const c=l("KButton"),v=l("KCard");return o(),r("div",g,[a(v,{ref:"emptyState","cta-is-hidden":"","is-error":!n.environment,class:"my-6"},b({body:i(()=>[n.environment==="kubernetes"?(o(),r("div",K,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",z,[R,e(),t("p",U,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),C,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",B,[E,e(),t("p",V,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):n.environment==="universal"?(o(),r("div",I,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",N,[W,e(),t("p",O,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",T,[j,e(),t("p",x,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):d("",!0)]),_:2},[n.environment==="kubernetes"||n.environment==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",S,w(n.environment),1)]),key:"0"}:void 0]),1032,["is-error"])])}const F=m(k,[["render",D],["__scopeId","data-v-6ac9df0a"]]);export{F as E};
