import{d as v,i as c,o as r,a as m,w as n,j as a,k as e,g as t,t as i,b as d,H as b,J as y,A as u,K as V,e as N,p as I,f as O,_ as k}from"./index-BmgyLxlv.js";import{m as x}from"./kong-icons.es265-Bx5prSMB.js";import{O as C,a as A,b as R}from"./OnboardingPage-SbSV0Kkq.js";const p=s=>(I("data-v-6d3e602f"),s=s(),O(),s),S=p(()=>t("strong",null,"few minutes",-1)),z={"data-testid":"kuma-environment"},B=p(()=>t("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),E={class:"item-status-list-wrapper"},L={class:"item-status-list"},T={class:"circle mr-2"},W=v({__name:"OnboardingWelcomeView",setup(s){return(K,j)=>{const _=c("RouteTitle"),g=c("AppView"),f=c("RouteView");return r(),m(f,{name:"onboarding-welcome-view"},{default:n(({env:h,t:o,can:w})=>[a(_,{title:o("onboarding.routes.welcome.title",{name:o("common.product.name")}),render:!1},null,8,["title"]),e(),a(g,null,{default:n(()=>[t("div",null,[a(C,null,{header:n(()=>[a(A,null,{title:n(()=>[e(`
                Welcome to `+i(o("common.product.name")),1)]),description:n(()=>[t("p",null,[e(`
                  Congratulations on downloading `+i(o("common.product.name"))+"! You are just a ",1),S,e(` away from getting your service mesh fully online.
                `)]),e(),t("p",null,[e(`
                  We have automatically detected that you are running on `),t("strong",z,i(o(`common.product.environment.${h("KUMA_ENVIRONMENT")}`)),1),e(`.
                `)])]),_:2},1024)]),content:n(()=>[B,e(),t("div",E,[t("ul",L,[(r(!0),d(b,null,y([{name:`Run ${o("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...w("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],l=>(r(),d("li",{key:l.name},[t("span",T,[l.status?(r(),m(u(x),{key:0,size:u(V)},null,8,["size"])):N("",!0)]),e(" "+i(l.name),1)]))),128))])])]),navigation:n(()=>[a(R,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}}),$=k(W,[["__scopeId","data-v-6d3e602f"]]);export{$ as default};
