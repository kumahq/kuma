import{d as h,C as M,i as m,j as U,o as A,c as R,a as n,w as o,e as a,g as t,u as e,E as c,M as i,y as _,N as f,O as C,H as v}from"./index-0be248c4.js";const l=r=>(f("data-v-70402eaf"),r=r(),C(),r),y={class:"kcard-list"},E=l(()=>t("p",null,[a(`
          We can create multiple isolated Mesh resources (i.e. per application/`),t("wbr"),a("team/"),t("wbr"),a(`business unit).
        `)],-1)),b={class:"resource-list-actions mt-4"},K=l(()=>t("p",null,`
          We need a data plane proxy for each replica of our services within a Mesh resource.
        `,-1)),g={class:"resource-list-actions mt-4"},T={class:"resource-list-actions mt-4"},k=["href"],S=["href"],x=["href"],D=h({__name:"MeshResources",setup(r){const s=M(),p=m(),u=U(()=>({name:p.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"}));return(d,N)=>(A(),R("div",y,[n(e(i),{title:"Create a virtual mesh"},{body:o(()=>[E,a(),t("div",b,[n(e(c),{icon:"plus",appearance:"creation",to:{name:"create-mesh"}},{default:o(()=>[a(`
            Create mesh
          `)]),_:1})])]),_:1}),a(),n(e(i),{title:"Connect data plane proxies"},{body:o(()=>[K,a(),t("div",g,[n(e(c),{to:e(u),appearance:"primary"},{default:o(()=>[a(`
            Get started
          `)]),_:1},8,["to"])])]),_:1}),a(),n(e(i),{title:`Apply ${e(s)("KUMA_PRODUCT_NAME")} policies`},{body:o(()=>[t("p",null,`
          We can apply `+_(e(s)("KUMA_PRODUCT_NAME"))+` policies to secure, observe, route and manage the Mesh and its data plane proxies.
        `,1),a(),t("div",T,[n(e(c),{to:`${e(s)("KUMA_DOCS_URL")}/policies/?${e(s)("KUMA_UTM_QUERY_PARAMS")}`,appearance:"primary",target:"_blank"},{default:o(()=>[a(`
            Explore policies
          `)]),_:1},8,["to"])])]),_:1},8,["title"]),a(),n(e(i),{title:"Resources"},{body:o(()=>[t("p",null,`
          Join the `+_(e(s)("KUMA_PRODUCT_NAME"))+` community and ask questions:
        `,1),a(),t("ul",null,[t("li",null,[t("a",{href:`${e(s)("KUMA_DOCS_URL")}/?${e(s)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(s)("KUMA_PRODUCT_NAME"))+` Documentation
            `,9,k)]),a(),t("li",null,[t("a",{href:`${e(s)("KUMA_CHAT_URL")}/?${e(s)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(s)("KUMA_PRODUCT_NAME"))+`  Community Chat
            `,9,S)]),a(),t("li",null,[t("a",{href:`https://github.com/kumahq/kuma?${e(s)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},_(e(s)("KUMA_PRODUCT_NAME"))+` GitHub Repository
            `,9,x)])])]),_:1})]))}});const O=v(D,[["__scopeId","data-v-70402eaf"]]);export{O as M};
