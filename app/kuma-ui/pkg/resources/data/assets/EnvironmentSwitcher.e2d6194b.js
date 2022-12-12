import{P as _,L as h,cn as p,E as m,i as l,o,j as r,a,K as w,w as i,e,l as t,t as y,A as d,C as b,D as f}from"./index.60b0f0ac.js";const g={name:"EnvironmentSwitcher",components:{KButton:_,KCard:h},data(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:{...p({environment:"config/getEnvironment"}),instructionsCtaText(){return this.environment==="universal"?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute(){return this.environment==="kubernetes"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}}},u=n=>(b("data-v-6ac9df0a"),n=n(),f(),n),k={class:"wizard-switcher"},K={class:"capitalize"},S={key:0},z={key:0},R=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),C={class:"text-center"},U=u(()=>t("br",null,null,-1)),E={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),I={class:"text-center"},N={key:1},V={key:0},W=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},P={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),j={class:"text-center"};function x(n,A,G,L,s,q){const c=l("KButton"),v=l("KCard");return o(),r("div",k,[a(v,{ref:"emptyState","cta-is-hidden":"","is-error":!n.environment,class:"my-6"},w({body:i(()=>[n.environment==="kubernetes"?(o(),r("div",S,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",z,[R,e(),t("p",C,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),U,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",E,[B,e(),t("p",I,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):n.environment==="universal"?(o(),r("div",N,[n.$route.name===s.wizardRoutes.kubernetes?(o(),r("div",V,[W,e(),t("p",D,[a(c,{to:{name:s.wizardRoutes.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n.$route.name===s.wizardRoutes.universal?(o(),r("div",P,[T,e(),t("p",j,[a(c,{to:{name:s.wizardRoutes.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):d("",!0)])):d("",!0)]),_:2},[n.environment==="kubernetes"||n.environment==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",K,y(n.environment),1)]),key:"0"}:void 0]),1032,["is-error"])])}const H=m(g,[["render",x],["__scopeId","data-v-6ac9df0a"]]);export{H as E};
