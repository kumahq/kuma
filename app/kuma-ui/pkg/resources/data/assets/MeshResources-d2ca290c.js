import{d as h,C as M,i as m,j as U,o as A,c as R,a as n,w as o,e as s,g as t,u as e,E as c,M as i,y as _,N as b,O as f,H as C}from"./index-a2028f71.js";const l=r=>(b("data-v-484d96b8"),r=r(),f(),r),v={class:"resource-list"},y=l(()=>t("p",null,[s(`
          We can create multiple isolated Mesh resources (i.e. per application/`),t("wbr"),s("team/"),t("wbr"),s(`business unit).
        `)],-1)),E={class:"resource-list-actions mt-4"},K=l(()=>t("p",null,`
          We need a data plane proxy for each replica of our services within a Mesh resource.
        `,-1)),g={class:"resource-list-actions mt-4"},T={class:"resource-list-actions mt-4"},S=["href"],k=["href"],x=["href"],D=h({__name:"MeshResources",setup(r){const a=M(),p=m(),u=U(()=>({name:p.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"}));return(d,N)=>(A(),R("div",v,[n(e(i),{title:"Create a virtual mesh"},{body:o(()=>[y,s(),t("div",E,[n(e(c),{icon:"plus",appearance:"creation",to:{name:"create-mesh"}},{default:o(()=>[s(`
            Create mesh
          `)]),_:1})])]),_:1}),s(),n(e(i),{title:"Connect data plane proxies"},{body:o(()=>[K,s(),t("div",g,[n(e(c),{to:e(u),appearance:"primary"},{default:o(()=>[s(`
            Get started
          `)]),_:1},8,["to"])])]),_:1}),s(),n(e(i),{title:`Apply ${e(a)("KUMA_PRODUCT_NAME")} policies`},{body:o(()=>[t("p",null,`
          We can apply `+_(e(a)("KUMA_PRODUCT_NAME"))+` policies to secure, observe, route and manage the Mesh and its data plane proxies.
        `,1),s(),t("div",T,[n(e(c),{to:`${e(a)("KUMA_DOCS_URL")}/policies/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,appearance:"primary",target:"_blank"},{default:o(()=>[s(`
            Explore policies
          `)]),_:1},8,["to"])])]),_:1},8,["title"]),s(),n(e(i),{title:"Resources"},{body:o(()=>[t("p",null,`
          Join the `+_(e(a)("KUMA_PRODUCT_NAME"))+` community and ask questions:
        `,1),s(),t("ul",null,[t("li",null,[t("a",{href:`${e(a)("KUMA_DOCS_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+` Documentation
            `,9,S)]),s(),t("li",null,[t("a",{href:`${e(a)("KUMA_CHAT_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+`  Community Chat
            `,9,k)]),s(),t("li",null,[t("a",{href:`https://github.com/kumahq/kuma?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(a)("KUMA_PRODUCT_NAME"))+` GitHub Repository
            `,9,x)])])]),_:1})]))}});const O=C(D,[["__scopeId","data-v-484d96b8"]]);export{O as M};
