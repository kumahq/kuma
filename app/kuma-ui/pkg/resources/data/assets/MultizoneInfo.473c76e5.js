import{k as i,ck as m,O as l,p as u,m as _,cc as t,o as d,i as K,w as n,a,b as e,e as s,bV as M,bW as $}from"./index.bd548025.js";const f=$(),g={name:"MultizoneInfo",env:f,productName:m,components:{KButton:l,KEmptyState:u,KIcon:_}},y=s("p",null,[e(`
        To access this page, you must be running in `),s("strong",null,"Multi-Zone"),e(` mode.
      `)],-1);function S(o,b,B,E,N,U){const c=t("KIcon"),p=t("KButton"),r=t("KEmptyState");return d(),K(r,null,{title:n(()=>[a(c,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),e(),s("p",null,M(o.$options.productName)+" is running in Standalone mode.",1)]),message:n(()=>[y]),cta:n(()=>[a(p,{to:`${o.$options.env("KUMA_DOCS_URL")}/documentation/deployments/?${o.$options.env("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank",appearance:"primary"},{default:n(()=>[e(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const v=i(g,[["render",S]]);export{v as M};
