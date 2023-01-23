import{m as _}from"./vuex.esm-bundler-df5bd11e.js";import{P as m,D as p}from"./kongponents.es-3df60cd6.js";import{_ as h}from"./_plugin-vue_export-helper-c27b6911.js";import{l,o as s,h as r,e as a,a0 as w,w as i,f as e,g as t,t as b,b as d,p as y,m as f}from"./runtime-dom.esm-bundler-91b41870.js";const g={name:"EnvironmentSwitcher",components:{KButton:m,KCard:p},data(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:{..._({environment:"config/getEnvironment"}),instructionsCtaText(){return this.environment==="universal"?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute(){return this.environment==="kubernetes"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}}},u=n=>(y("data-v-6ac9df0a"),n=n(),f(),n),k={class:"wizard-switcher"},S={class:"capitalize"},K={key:0},z={key:0},R=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),C={class:"text-center"},U=u(()=>t("br",null,null,-1)),B={key:1},E=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),I={class:"text-center"},N={key:1},V={key:0},W=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},P={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),x={class:"text-center"};function G(n,j,q,A,o,F){const c=l("KButton"),v=l("KCard");return s(),r("div",k,[a(v,{ref:"emptyState","cta-is-hidden":"","is-error":!n.environment,class:"my-6"},w({body:i(()=>[n.environment==="kubernetes"?(s(),r("div",K,[n.$route.name===o.wizardRoutes.kubernetes?(s(),r("div",z,[R,e(),t("p",C,[a(c,{to:{name:o.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),U,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===o.wizardRoutes.universal?(s(),r("div",B,[E,e(),t("p",I,[a(c,{to:{name:o.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):n.environment==="universal"?(s(),r("div",N,[n.$route.name===o.wizardRoutes.kubernetes?(s(),r("div",V,[W,e(),t("p",D,[a(c,{to:{name:o.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===o.wizardRoutes.universal?(s(),r("div",P,[T,e(),t("p",x,[a(c,{to:{name:o.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):d("",!0)]),_:2},[n.environment==="kubernetes"||n.environment==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",S,b(n.environment),1)]),key:"0"}:void 0]),1032,["is-error"])])}const O=h(g,[["render",G],["__scopeId","data-v-6ac9df0a"]]);export{O as E};
