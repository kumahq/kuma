import{d as h,D as M,i as m,j as U,o as A,c as R,a as n,w as o,e as s,g as t,u as e,E as i,M as c,y as _,N as b,O as f,_ as v}from"./index-3d59543a.js";const l=r=>(b("data-v-88c956eb"),r=r(),f(),r),y={class:"kcard-switcher"},C=l(()=>t("p",null,[s(`
          We can create multiple isolated Mesh resources (i.e. per application/`),t("wbr"),s("team/"),t("wbr"),s(`business unit).
        `)],-1)),E={class:"resource-list-actions mt-4"},K=l(()=>t("p",null,`
          We need a data plane proxy for each replica of our services within a Mesh resource.
        `,-1)),g={class:"resource-list-actions mt-4"},T={class:"resource-list-actions mt-4"},k=["href"],D=["href"],S=["href"],x=h({__name:"MeshResources",setup(r){const a=M(),p=m(),u=U(()=>({name:p.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"}));return(d,N)=>(A(),R("div",y,[n(e(c),{title:"Create a virtual mesh"},{body:o(()=>[C,s(),t("div",E,[n(e(i),{icon:"plus",appearance:"creation",to:{name:"create-mesh"}},{default:o(()=>[s(`
            Create mesh
          `)]),_:1})])]),_:1}),s(),n(e(c),{title:"Connect data plane proxies"},{body:o(()=>[K,s(),t("div",g,[n(e(i),{to:e(u),appearance:"primary"},{default:o(()=>[s(`
            Get started
          `)]),_:1},8,["to"])])]),_:1}),s(),n(e(c),{title:`Apply ${e(a)("KUMA_PRODUCT_NAME")} policies`},{body:o(()=>[t("p",null,`
          We can apply `+_(e(a)("KUMA_PRODUCT_NAME"))+` policies to secure, observe, route and manage the Mesh and its data plane proxies.
        `,1),s(),t("div",T,[n(e(i),{to:`${e(a)("KUMA_DOCS_URL")}/policies/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,appearance:"primary",target:"_blank"},{default:o(()=>[s(`
            Explore policies
          `)]),_:1},8,["to"])])]),_:1},8,["title"]),s(),n(e(c),{title:"Resources"},{body:o(()=>[t("p",null,`
          Join the `+_(e(a)("KUMA_PRODUCT_NAME"))+` community and ask questions:
        `,1),s(),t("ul",null,[t("li",null,[t("a",{href:`${e(a)("KUMA_DOCS_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+` Documentation
            `,9,k)]),s(),t("li",null,[t("a",{href:`${e(a)("KUMA_CHAT_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+`  Community Chat
            `,9,D)]),s(),t("li",null,[t("a",{href:`https://github.com/kumahq/kuma?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+` GitHub Repository
            `,9,S)])])]),_:1})]))}});const O=v(x,[["__scopeId","data-v-88c956eb"]]);export{O as M};
