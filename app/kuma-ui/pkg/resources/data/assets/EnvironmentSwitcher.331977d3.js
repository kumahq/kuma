import{A as _,O as p,co as h,H as m,i as l,o,j as r,a,M as w,w as i,b as e,l as t,t as b,B as d,E as y,G as f}from"./index.e014f0d3.js";const g={name:"EnvironmentSwitcher",components:{KButton:_,KCard:p},data(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:{...h({environment:"config/getEnvironment"}),instructionsCtaText(){return this.environment==="universal"?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute(){return this.environment==="kubernetes"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}}},u=n=>(y("data-v-6ac9df0a"),n=n(),f(),n),k={class:"wizard-switcher"},S={class:"capitalize"},K={key:0},z={key:0},R=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),C={class:"text-center"},U=u(()=>t("br",null,null,-1)),B={key:1},E=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),I={class:"text-center"},N={key:1},V={key:0},W=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},G={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),j={class:"text-center"};function x(n,D,H,M,s,O){const c=l("KButton"),v=l("KCard");return o(),r("div",k,[a(v,{ref:"emptyState","cta-is-hidden":"","is-error":!n.environment,class:"my-6"},w({body:i(()=>[n.environment==="kubernetes"?(o(),r("div",K,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",z,[R,e(),t("p",C,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),U,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",B,[E,e(),t("p",I,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):n.environment==="universal"?(o(),r("div",N,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",V,[W,e(),t("p",A,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",G,[T,e(),t("p",j,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):d("",!0)]),_:2},[n.environment==="kubernetes"||n.environment==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",S,b(n.environment),1)]),key:"0"}:void 0]),1032,["is-error"])])}const F=m(g,[["render",x],["__scopeId","data-v-6ac9df0a"]]);export{F as E};
