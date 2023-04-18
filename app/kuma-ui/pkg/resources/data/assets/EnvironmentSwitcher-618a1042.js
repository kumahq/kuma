import{d as p,l as m,h as y,i as b,o,c as a,a as c,Q as f,u as n,w as i,e,f as t,z as g,_ as l,y as v,I as k,N as w,O as S,J as x}from"./index-5f1fbf13.js";const u=d=>(w("data-v-f74b1174"),d=d(),S(),d),K={class:"wizard-switcher"},U={class:"capitalize"},z={key:0},E={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},D={key:0},J=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),O={class:"text-center"},Q={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),j={class:"text-center"},q=p({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=m(),h=y(),r=b(()=>h.getters["config/getEnvironment"]);return(A,F)=>(o(),a("div",K,[c(n(k),{ref:"emptyState","cta-is-hidden":"","is-error":!n(r),class:"my-6"},f({body:i(()=>[n(r)==="kubernetes"?(o(),a("div",z,[n(_).name===s.kubernetes?(o(),a("div",E,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):n(r)==="universal"?(o(),a("div",R,[n(_).name===s.kubernetes?(o(),a("div",D,[J,e(),t("p",O,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",Q,[T,e(),t("p",j,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[n(r)==="kubernetes"||n(r)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",U,g(n(r)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const H=x(q,[["__scopeId","data-v-f74b1174"]]);export{H as E};
